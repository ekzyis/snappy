package client_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	sn "github.com/ekzyis/snappy"
)

// TestFixturesRoundTrip decodes each mock fixture into its response type and
// checks two things:
//
//   - representative non-null fields are populated, so a mistyped json tag on
//     one of them (e.g. a `parentId` field tagged `parentIdx`) is caught;
//   - re-encoding the decoded value and decoding it again yields an identical
//     encoding, i.e. the type is a stable fixed point over (un)marshaling.
func TestFixturesRoundTrip(t *testing.T) {
	fx := fixtureData

	t.Run("items", func(t *testing.T) {
		var resp sn.ItemsResponse
		decodeData(t, fx["items"], "items", &resp)

		if len(resp.Data.Items.Items) == 0 {
			t.Fatal("no items decoded")
		}
		it := resp.Data.Items.Items[0]
		if it.Id == 0 {
			t.Error("item id not populated")
		}
		if it.User.Id == 0 {
			t.Error("item user id not populated")
		}
		if it.CreatedAt.IsZero() {
			t.Error("item createdAt not populated")
		}
		assertIdempotent(t, resp)
	})

	t.Run("item", func(t *testing.T) {
		var resp sn.ItemResponse
		decodeData(t, fx["item"], "item", &resp)

		if resp.Data.Item.Id == 0 {
			t.Error("item id not populated")
		}
		if resp.Data.Item.CreatedAt.IsZero() {
			t.Error("item createdAt not populated")
		}
		assertIdempotent(t, resp)
	})

	t.Run("dupes", func(t *testing.T) {
		var resp sn.DupesResponse
		decodeData(t, fx["dupes"], "dupes", &resp)

		if len(resp.Data.Dupes) == 0 {
			t.Fatal("no dupes decoded")
		}
		d := resp.Data.Dupes[0]
		if d.Id == 0 {
			t.Error("dupe id not populated")
		}
		if d.Url == "" {
			t.Error("dupe url not populated")
		}
		assertIdempotent(t, resp)
	})

	t.Run("me", func(t *testing.T) {
		var resp sn.MeResponse
		decodeData(t, fx["me"], "me", &resp)

		if resp.Data.Me.Id == 0 {
			t.Error("me id not populated")
		}
		if resp.Data.Me.Name == "" {
			t.Error("me name not populated")
		}
		assertIdempotent(t, resp)
	})

	t.Run("notifications", func(t *testing.T) {
		var resp sn.NotificationsResponse
		decodeData(t, fx["notifications"], "notifications", &resp)
		assertIdempotent(t, resp)
	})

	t.Run("upsertDiscussion", func(t *testing.T) {
		var resp sn.UpsertDiscussionResponse
		decodeData(t, fx["upsertDiscussion"], "upsertDiscussion", &resp)

		if resp.Data.UpsertDiscussion.PayerPrivates.Result.Id == 0 {
			t.Error("payIn result id not populated")
		}
		assertIdempotent(t, resp)
	})

	t.Run("getSignedPOST", func(t *testing.T) {
		var resp sn.GetSignedPOSTResponse
		decodeData(t, fx["getSignedPOST"], "getSignedPOST", &resp)

		if resp.Data.GetSignedPOST.Url == "" {
			t.Error("signed post url not populated")
		}
		assertIdempotent(t, resp)
	})
}

// decodeData wraps a fixture as a full GraphQL response and decodes it into out.
func decodeData(t *testing.T, raw json.RawMessage, field string, out any) {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"data": map[string]json.RawMessage{field: raw},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, out); err != nil {
		t.Fatalf("decode %s: %v", field, err)
	}
}

// assertIdempotent verifies that encoding v, decoding it back into the same
// type, and re-encoding produces identical bytes.
func assertIdempotent(t *testing.T, v any) {
	t.Helper()
	first, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	again := reflect.New(reflect.TypeOf(v)).Interface()
	if err := json.Unmarshal(first, again); err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	second, err := json.Marshal(reflect.ValueOf(again).Elem().Interface())
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("not idempotent over (un)marshal:\n first: %s\nsecond: %s", first, second)
	}
}
