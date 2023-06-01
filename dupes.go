package sn

import (
	"encoding/json"
	"fmt"
)

// Fetch dupes
func Dupes(url string) (*[]Dupe, error) {
	body := GraphQLPayload{
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
	resp, err := MakeStackerNewsRequest(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody DupesResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding SN dupes: %w", err)
		return nil, err
	}
	err = CheckForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}

	return &respBody.Data.Dupes, nil
}
