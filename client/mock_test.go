package client_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sn "github.com/ekzyis/snappy"
	types "github.com/ekzyis/snappy/types"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

//go:generate sh -c "cd ../scripts && npm install --no-audit --no-fund && npm run dump-schema"

var (
	// schema is the real Stacker News GraphQL schema, dumped from the submodule
	// into testdata/schema.graphql. Regenerate with `make schema`.
	schema *ast.Schema
	// fixtureData holds the mock responses keyed by operation field, loaded from
	// testdata/fixtures/.
	fixtureData map[string]json.RawMessage
)

func TestMain(m *testing.M) {
	src, err := os.ReadFile(filepath.Join("testdata", "schema.graphql"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "read schema: %v\n", err)
		os.Exit(1)
	}
	var errs error
	schema, errs = gqlparser.LoadSchema(&ast.Source{Name: "schema.graphql", Input: string(src)})
	if errs != nil {
		fmt.Fprintf(os.Stderr, "load schema: %v\n", errs)
		os.Exit(1)
	}

	fixtureData, err = loadFixtures()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load fixtures: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// newTestClient returns a client pointed at an in-process mock server. The
// server validates every outgoing GraphQL query against the real SN schema and
// replies with a canned fixture for the operation, so tests run with no network
// and no live SN instance.
func newTestClient(t *testing.T) *sn.Client {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Non-GraphQL paths (e.g. the S3 upload in UploadImage) just succeed.
		if r.URL.Path != "/api/graphql" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var body types.GqlBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeGqlError(w, "decode request: "+err.Error())
			return
		}

		field, err := validateQuery(body.Query)
		if err != nil {
			writeGqlError(w, err.Error())
			return
		}

		fixture, ok := fixtureData[field]
		if !ok {
			writeGqlError(w, "no fixture for operation: "+field)
			return
		}

		writeJSON(w, map[string]any{"data": map[string]json.RawMessage{field: fixture}})
	})

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return sn.NewClient(
		sn.WithBaseUrl(srv.URL),
		sn.WithApiKey("test"),
		sn.WithMediaUrl(srv.URL),
	)
}

// validateQuery checks a query against the schema and returns the name of its
// top-level field, which selects the fixture to reply with.
func validateQuery(query string) (string, error) {
	doc, errs := gqlparser.LoadQueryWithRules(schema, query, rules.NewDefaultRules())
	if errs != nil {
		return "", errs
	}
	if len(doc.Operations) == 0 || len(doc.Operations[0].SelectionSet) == 0 {
		return "", fmt.Errorf("query has no top-level field")
	}
	field, ok := doc.Operations[0].SelectionSet[0].(*ast.Field)
	if !ok {
		return "", fmt.Errorf("top-level selection is not a field")
	}
	return field.Name, nil
}

// loadFixtures reads every testdata/fixtures/<field>.json into a map keyed by
// operation field.
func loadFixtures() (map[string]json.RawMessage, error) {
	matches, err := filepath.Glob(filepath.Join("testdata", "fixtures", "*.json"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no fixtures in testdata/fixtures; run `make fixtures`")
	}

	fx := make(map[string]json.RawMessage, len(matches))
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		field := strings.TrimSuffix(filepath.Base(path), ".json")
		fx[field] = json.RawMessage(data)
	}

	return fx, nil
}

func writeGqlError(w http.ResponseWriter, msg string) {
	writeJSON(w, map[string]any{"errors": []types.GqlError{{Message: msg}}})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
