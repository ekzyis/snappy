package sn

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Id       int          `json:"id,string"`
	Name     string       `json:"name"`
	Privates UserPrivates `json:"privates"`
}

type UserPrivates struct {
	Sats int `json:"sats"`
}

type MeResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		Me User `json:"me"`
	} `json:"data"`
}

func (c *Client) Me() (*User, error) {
	body := GqlBody{
		Query: `
		query me {
			me {
				id
				name
				privates {
					sats
				}
			}
		}`,
	}

	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody MeResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding me: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}
	return &respBody.Data.Me, nil
}
