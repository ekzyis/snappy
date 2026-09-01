package client_test

import (
	"testing"
)

func TestQueryItems(t *testing.T) {
	c := newTestClient(t)

	cursor, err := c.Items(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cursor.Items) == 0 {
		t.Fatal("items cursor empty")
	}
}

func TestQueryItem(t *testing.T) {
	c := newTestClient(t)

	item, err := c.Item(1)
	if err != nil {
		t.Fatal(err)
	}
	if item.Id == 0 {
		t.Fatal("item id missing")
	}
}

func TestQueryDupes(t *testing.T) {
	c := newTestClient(t)

	dupes, err := c.Dupes("https://stacker.news")
	if err != nil {
		t.Fatal(err)
	}
	if len(*dupes) == 0 {
		t.Fatal("dupes empty")
	}
}

func TestQueryMe(t *testing.T) {
	c := newTestClient(t)

	me, err := c.Me()
	if err != nil {
		t.Fatal(err)
	}
	if me.Name == "" {
		t.Fatal("me name missing")
	}
}

func TestQueryNotifications(t *testing.T) {
	c := newTestClient(t)

	if _, err := c.Notifications(); err != nil {
		t.Fatal(err)
	}
}

func TestMutationCreateComment(t *testing.T) {
	c := newTestClient(t)

	if _, err := c.CreateComment(1, "test comment"); err != nil {
		t.Fatal(err)
	}
}

func TestMutationPostDiscussion(t *testing.T) {
	c := newTestClient(t)

	if _, err := c.PostDiscussion("test discussion", "test discussion text", []string{"bitcoin"}); err != nil {
		t.Fatal(err)
	}
}

func TestMutationPostLink(t *testing.T) {
	c := newTestClient(t)

	if _, err := c.PostLink("https://stacker.news", "test link", "test link text", []string{"bitcoin"}); err != nil {
		t.Fatal(err)
	}
}
