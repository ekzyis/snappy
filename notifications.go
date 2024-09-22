package sn

import (
	"encoding/json"
	"fmt"
	"time"
)

type Notification struct {
	Id   int    `json:"id,string"`
	Type string `json:"__typename"`
	Item Item   `json:"item"`
}

type NotificationsCursor struct {
	LastChecked   time.Time      `json:"lastChecked"`
	Cursor        string         `json:"cursor"`
	Notifications []Notification `json:"notifications"`
}

type NotificationsResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		Notifications NotificationsCursor `json:"notifications"`
	} `json:"data"`
}

func (c *Client) Notifications() (*NotificationsCursor, error) {
	body := GqlBody{
		Query: `
		fragment ItemFields on Item {
			id
			user {
				id
				name
			}
			parentId
			createdAt
			deletedAt
			title
			text
		}
		query notifications {
			notifications {
				lastChecked
				cursor
				notifications {
					__typename
					... on Reply {
						id
						item {
							...ItemFields
						}
					}
					... on Mention {
						id
						item {
							...ItemFields
						}
					}
				}
			}
		}
		`,
		Variables: map[string]interface{}{},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody NotificationsResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding notifications: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}
	return &respBody.Data.Notifications, nil
}

func (c *Client) Mentions() ([]Notification, error) {
	return c.filterNotifications(
		func(n Notification) bool {
			return n.Type == "Mention"
		},
	)
}

func (c *Client) filterNotifications(f func(Notification) bool) ([]Notification, error) {
	var (
		n   *NotificationsCursor
		err error
	)

	if n, err = c.Notifications(); err != nil {
		return nil, err
	}

	return filter(n.Notifications, f), nil
}

func filter[T any](s []T, f func(T) bool) []T {
	var r []T
	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}
