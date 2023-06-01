package sn

import (
	"encoding/json"
	"fmt"
)

func PostLink(url string, title string, sub string) (int, error) {
	body := GraphQLPayload{
		Query: `
	 		mutation upsertLink($url: String!, $title: String!, $sub: String!) {
	 			upsertLink(url: $url, title: $title, sub: $sub) {
	 				id
	 			}
	 		}`,
		Variables: map[string]interface{}{
			"url":   url,
			"title": title,
			"sub":   sub,
		},
	}
	resp, err := MakeStackerNewsRequest(body)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var respBody UpsertLinkResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding SN upsertLink: %w", err)
		return -1, err
	}
	err = CheckForErrors(respBody.Errors)
	if err != nil {
		return -1, err
	}
	itemId := respBody.Data.UpsertLink.Id
	return itemId, nil
}

func CreateComment(parentId int, text string) (int, error) {
	body := GraphQLPayload{
		Query: `
			mutation createComment($text: String!, $parentId: ID!) {
        createComment(text: $text, parentId: $parentId) {
          id
        }
			}`,
		Variables: map[string]interface{}{
			"text":     text,
			"parentId": parentId,
		},
	}
	resp, err := MakeStackerNewsRequest(body)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var respBody CreateCommentsResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding SN createComment: %w", err)
		return -1, err
	}
	err = CheckForErrors(respBody.Errors)
	if err != nil {
		return -1, err
	}

	return parentId, nil
}
