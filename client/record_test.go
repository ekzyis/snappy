//go:build record

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testNsec is a throwaway Nostr identity used only to record fixtures against
// the live Stacker News API. 't' is replaced with 'n' before the HTTP request.
const testNsec = "tsec17qfg2qxuhh8e2n35zhfejv5e37jdzswd24fytltcvnx6qcmf0rssgwe66n"

// recordFields are the operations captured from the live API. The login flow
// also calls the API (createAuth), so we only keep these.
var recordFields = map[string]bool{
	"items":         true,
	"item":          true,
	"dupes":         true,
	"me":            true,
	"notifications": true,
}

// TestRecordFixtures makes live calls to the real Stacker News API, snapshots
// each raw response under testdata/live/<field>.json, and verifies that the
// committed fixtures in testdata/fixtures/ still match the live API shape (see
// verifyShapes). It logs in with testNsec so authenticated reads (me,
// notifications) are covered too. Run it with `make fixtures`.
//
// It is excluded from normal `go test` by the `record` build tag so the test
// suite stays hermetic.
func TestRecordFixtures(t *testing.T) {
	dir := filepath.Join("testdata", "live")
	if err := os.MkdirAll(dir, 0o775); err != nil {
		t.Fatal(err)
	}

	rec := &fixtureRecorder{next: http.DefaultTransport, dir: dir, t: t, got: map[string]json.RawMessage{}}
	c := NewClient(WithNsec(strings.Replace(testNsec, "t", "n", 1)))
	c.httpClient.Transport = rec

	cursor, err := c.Items(nil)
	if err != nil {
		t.Fatalf("items: %v", err)
	}
	if len(cursor.Items) == 0 {
		t.Fatal("items returned nothing; cannot bootstrap item/dupes fixtures")
	}

	if _, err := c.Item(cursor.Items[0].Id); err != nil {
		t.Fatalf("item: %v", err)
	}

	// dupes(url) returns items posted with that url, so an existing item's url
	// guarantees a non-empty response.
	for _, it := range cursor.Items {
		if it.Url != "" {
			if _, err := c.Dupes(it.Url); err != nil {
				t.Fatalf("dupes: %v", err)
			}
			break
		}
	}

	if _, err := c.Me(); err != nil {
		t.Fatalf("me: %v", err)
	}

	if _, err := c.Notifications(); err != nil {
		t.Fatalf("notifications: %v", err)
	}

	verifyShapes(t, rec.got)
}

// verifyShapes checks that every committed fixture is shape-compatible with the
// live response for the same operation: each field the fixture hardcodes must
// appear in the live response with a compatible type. It compares structure and
// types only, never values. notifications is polymorphic (its entries vary by
// account), so only its envelope is checked.
func verifyShapes(t *testing.T, got map[string]json.RawMessage) {
	for field := range recordFields {
		live, ok := got[field]
		if !ok {
			t.Errorf("%s: no live response captured", field)
			continue
		}

		fixture, err := os.ReadFile(filepath.Join("testdata", "fixtures", field+".json"))
		if err != nil {
			t.Errorf("%s: read fixture: %v", field, err)
			continue
		}

		var want, have any
		if err := json.Unmarshal(fixture, &want); err != nil {
			t.Errorf("%s: decode fixture: %v", field, err)
			continue
		}
		if err := json.Unmarshal(live, &have); err != nil {
			t.Errorf("%s: decode live response: %v", field, err)
			continue
		}

		var diffs []string
		if field == "notifications" {
			diffs = checkEnvelope(have)
		} else {
			diffs = checkShape(field, want, have)
		}
		for _, d := range diffs {
			t.Errorf("fixture out of sync with API: %s", d)
		}
	}
}

// checkShape reports where the fixture (want) expects a field or type the live
// response (have) does not provide. A null on either side is compatible, since
// any field may be null. Arrays are compared against the first live element.
func checkShape(path string, want, have any) []string {
	if want == nil || have == nil {
		return nil
	}
	switch w := want.(type) {
	case map[string]any:
		h, ok := have.(map[string]any)
		if !ok {
			return []string{fmt.Sprintf("%s: fixture is object, API returned %s", path, typeName(have))}
		}
		var diffs []string
		for k, wv := range w {
			hv, present := h[k]
			if !present {
				diffs = append(diffs, fmt.Sprintf("%s.%s: not present in API response", path, k))
				continue
			}
			diffs = append(diffs, checkShape(path+"."+k, wv, hv)...)
		}
		return diffs
	case []any:
		h, ok := have.([]any)
		if !ok {
			return []string{fmt.Sprintf("%s: fixture is array, API returned %s", path, typeName(have))}
		}
		if len(h) == 0 {
			return nil
		}
		var diffs []string
		for i, wv := range w {
			diffs = append(diffs, checkShape(fmt.Sprintf("%s[%d]", path, i), wv, h[0])...)
		}
		return diffs
	default:
		if typeName(want) != typeName(have) {
			return []string{fmt.Sprintf("%s: fixture is %s, API returned %s", path, typeName(want), typeName(have))}
		}
		return nil
	}
}

// checkEnvelope verifies the top-level shape of a notifications response without
// descending into its polymorphic entries.
func checkEnvelope(have any) []string {
	h, ok := have.(map[string]any)
	if !ok {
		return []string{fmt.Sprintf("notifications: fixture is object, API returned %s", typeName(have))}
	}
	var diffs []string
	for _, k := range []string{"lastChecked", "cursor", "notifications"} {
		if _, present := h[k]; !present {
			diffs = append(diffs, "notifications."+k+": not present in API response")
		}
	}
	if _, ok := h["notifications"].([]any); !ok && h["notifications"] != nil {
		diffs = append(diffs, "notifications.notifications: not an array")
	}
	return diffs
}

func typeName(v any) string {
	switch v.(type) {
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "bool"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%T", v)
	}
}

// fixtureRecorder tees the raw data payload of each live GraphQL response to disk
// and keeps it for shape verification, keyed by the response's top-level field.
type fixtureRecorder struct {
	next http.RoundTripper
	dir  string
	t    *testing.T
	got  map[string]json.RawMessage
}

func (r *fixtureRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := r.next.RoundTrip(req)
	if err != nil || resp == nil || req.URL.Path != "/api/graphql" {
		return resp, err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	var parsed struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return resp, nil
	}

	for field, val := range parsed.Data {
		if !recordFields[field] {
			continue
		}
		r.got[field] = val

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, val, "", "  "); err != nil {
			pretty.Write(val)
		}
		pretty.WriteByte('\n')
		path := filepath.Join(r.dir, field+".json")
		if err := os.WriteFile(path, pretty.Bytes(), 0o644); err != nil {
			r.t.Errorf("write %s: %v", path, err)
			continue
		}
		r.t.Logf("recorded %s", path)
	}

	return resp, nil
}
