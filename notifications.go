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
