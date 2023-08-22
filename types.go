package sn

import (
	"fmt"
	"time"
)

type GraphQLPayload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

type User struct {
	Id   int    `json:"id,string"`
	Name string `json:"name"`
}

type Comment struct {
	Id       int       `json:"id,string"`
	Text     string    `json:"text"`
	User     User      `json:"user"`
	Comments []Comment `json:"comments"`
}

type CreateCommentsResponse struct {
	Errors []GraphQLError `json:"errors"`
	Data   struct {
		CreateComment Comment `json:"createComment"`
	} `json:"data"`
}

type Item struct {
	Id        int       `json:"id,string"`
	ParentId  int       `json:"parentId,string"`
	Title     string    `json:"title"`
	Url       string    `json:"url"`
	Sats      int       `json:"sats"`
	CreatedAt time.Time `json:"createdAt"`
	Comments  []Comment `json:"comments"`
	NComments int       `json:"ncomments"`
	User      User      `json:"user"`
}

type UpsertLinkResponse struct {
	Errors []GraphQLError `json:"errors"`
	Data   struct {
		UpsertLink Item `json:"upsertLink"`
	} `json:"data"`
}

type ItemsResponse struct {
	Errors []GraphQLError `json:"errors"`
	Data   struct {
		Items ItemsCursor `json:"items"`
	} `json:"data"`
}

type ItemsCursor struct {
	Items  []Item `json:"items"`
	Cursor string `json:"cursor"`
}

type HasNewNotesResponse struct {
	Errors []GraphQLError `json:"errors"`
	Data   struct {
		HasNewNotes bool `json:"hasNewNotes"`
	} `json:"data"`
}

type Dupe struct {
	Id        int       `json:"id,string"`
	Url       string    `json:"url"`
	Title     string    `json:"title"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"createdAt"`
	Sats      int       `json:"sats"`
	NComments int       `json:"ncomments"`
}

type DupesResponse struct {
	Errors []GraphQLError `json:"errors"`
	Data   struct {
		Dupes []Dupe `json:"dupes"`
	} `json:"data"`
}

type DupesError struct {
	Url   string
	Dupes []Dupe
}

func (e *DupesError) Error() string {
	return fmt.Sprintf("found %d dupes for %s", len(e.Dupes), e.Url)
}

type RssItem struct {
	Guid        string    `xml:"guid"`
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     RssDate   `xml:"pubDate"`
	Author      RssAuthor `xml:"author"`
}

type RssChannel struct {
	Title         string    `xml:"title"`
	Description   string    `xml:"description"`
	Link          string    `xml:"link"`
	Items         []RssItem `xml:"item"`
	LastBuildDate RssDate   `xml:"lastBuildDate"`
}

type Rss struct {
	Channel RssChannel `xml:"channel"`
}

type RssDate struct {
	time.Time
}

type RssAuthor struct {
	Name string `xml:"name"`
}

type ItemsQuery struct {
	Sub    string
	Sort   string
	Type   string
	Cursor string
	Name   string
	When   string
	By     string
	Limit  int
}
