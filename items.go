package sn

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/guregu/null.v4"
)

type Item struct {
	Id        int       `json:"id,string"`
	ParentId  int       `json:"parentId"`
	Title     string    `json:"title"`
	Url       string    `json:"url"`
	Text      string    `json:"text"`
	Sats      int       `json:"sats"`
	CreatedAt time.Time `json:"createdAt"`
	DeletedAt null.Time `json:"deletedAt"`
	Comments  []Comment `json:"comments"`
	NComments int       `json:"ncomments"`
	User      User      `json:"user"`
}

type Comment struct {
	Id        int       `json:"id,string"`
	ParentId  int       `json:"parentId"`
	CreatedAt time.Time `json:"createdAt"`
	Text      string    `json:"text"`
	User      User      `json:"user"`
	Comments  []Comment `json:"comments"`
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

type ItemsCursor struct {
	Items  []Item `json:"items"`
	Cursor string `json:"cursor"`
}

type ItemResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		Item Item `json:"item"`
	} `json:"data"`
}

type ItemsResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		Items ItemsCursor `json:"items"`
	} `json:"data"`
}

type ItemPaidAction struct {
	Result        Item          `json:"result"`
	Invoice       Invoice       `json:"invoice"`
	PaymentMethod PaymentMethod `json:"paymentMethod"`
}

type UpsertDiscussionResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		UpsertDiscussion ItemPaidAction `json:"upsertDiscussion"`
	} `json:"data"`
}

type UpsertLinkResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		UpsertLink ItemPaidAction `json:"upsertLink"`
	} `json:"data"`
}

type UpsertCommentResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		UpsertComment ItemPaidAction `json:"upsertComment"`
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
	Errors []GqlError `json:"errors"`
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

func (c *Client) Item(id int) (*Item, error) {
	body := GqlBody{
		Query: `
		query item($id: ID!) {
			item(id: $id) {
				id
				parentId
				title
				url
				text
				sats
				createdAt
				deletedAt
				ncomments
				user {
					id
					name
				}
			}
		}`,
		Variables: map[string]interface{}{
			"id": id,
		},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody ItemResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding item: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}
	return &respBody.Data.Item, nil
}

func (c *Client) Items(query *ItemsQuery) (*ItemsCursor, error) {
	if query == nil {
		query = &ItemsQuery{}
	}

	body := GqlBody{
		Query: `
		query items($sub: String, $sort: String, $cursor: String, $type: String, $name: String, $when: String, $by: String, $limit: Limit) {
			items(sub: $sub, sort: $sort, cursor: $cursor, type: $type, name: $name, when: $when, by: $by, limit: $limit) {
				cursor
				items {
					id
					parentId
					title
					url
					text
					sats
					createdAt
					deletedAt
					ncomments
					user {
						id
						name
					}
				},
			}
		}`,
		Variables: map[string]interface{}{
			"sub":    query.Sub,
			"sort":   query.Sort,
			"type":   query.Type,
			"cursor": query.Cursor,
			"name":   query.Name,
			"when":   query.When,
			"by":     query.By,
			"limit":  query.Limit,
		},
	}
	if query.Limit == 0 {
		body.Variables["limit"] = 21
	}

	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody ItemsResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding items: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}
	return &respBody.Data.Items, nil
}

func (c *Client) PostDiscussion(title string, text string, sub string) (int, error) {
	body := GqlBody{
		Query: `
		mutation upsertDiscussion($title: String!, $text: String, $sub: String) {
			upsertDiscussion(title: $title, text: $text, sub: $sub) {
				result { id }
			}
		}`,
		Variables: map[string]interface{}{
			"title": title,
			"text":  text,
			"sub":   sub,
		},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var respBody UpsertDiscussionResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding upsertDiscussion: %w", err)
		return -1, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return -1, err
	}

	return respBody.Data.UpsertDiscussion.Result.Id, nil
}

func (c *Client) PostLink(url string, title string, text string, sub string) (int, error) {
	body := GqlBody{
		Query: `
		mutation upsertLink($url: String!, $title: String!, $text: String, $sub: String!) {
			upsertLink(url: $url, title: $title, text: $text, sub: $sub) {
				result { id }
			}
		}`,
		Variables: map[string]interface{}{
			"url":   url,
			"title": title,
			"text":  text,
			"sub":   sub,
		},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var respBody UpsertLinkResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding upsertLink: %w", err)
		return -1, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return -1, err
	}

	return respBody.Data.UpsertLink.Result.Id, nil
}

func (c *Client) CreateComment(parentId int, text string) (int, error) {
	body := GqlBody{
		Query: `
		mutation upsertComment($parentId: ID!, $text: String!) {
			upsertComment(parentId: $parentId, text: $text) {
				result { id }
			}
		}`,
		Variables: map[string]interface{}{
			"parentId": parentId,
			"text":     text,
		},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var respBody UpsertCommentResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding upsertComment: %w", err)
		return -1, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return -1, err
	}

	return respBody.Data.UpsertComment.Result.Id, nil
}

func (c *Client) Dupes(url string) (*[]Dupe, error) {
	body := GqlBody{
		Query: `
		query Dupes($url: String!) {
			dupes(url: $url) {
				id
				url
				title
				user {
					name
				}
				createdAt
				sats
				ncomments
			}
		}`,
		Variables: map[string]interface{}{
			"url": url,
		},
	}
	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody DupesResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding dupes: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}

	return &respBody.Data.Dupes, nil
}

func (c *Client) HasDupes(url string) (bool, error) {
	dupes, err := c.Dupes(url)
	if err != nil {
		return false, err
	}

	return len(*dupes) > 0, nil
}
