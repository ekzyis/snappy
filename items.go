package sn

import (
	"encoding/json"
	"fmt"
)

func Items(query *ItemsQuery) (*ItemsCursor, error) {
	if query == nil {
		query = &ItemsQuery{}
	}

	if sub := query.Sub; sub != "" {
		if !(sub == "bitcoin" || sub == "nostr" || sub == "tech" || sub == "meta") {
			return nil, fmt.Errorf("invalid sub: %s", sub)
		}
	}

	body := GraphQLPayload{
		Query: `
			query items($sub: String, $sort: String, $cursor: String, $type: String, $name: String, $when: String, $by: String, $limit: Limit) {
				items(sub: $sub, sort: $sort, cursor: $cursor, type: $type, name: $name, when: $when, by: $by, limit: $limit) {
					cursor
					items {
						id
						parentId
						createdAt
						deletedAt
						title
						url
						user {
							id
							name
						}
						otsHash
						position
						sats
						boost
						bounty
						bountyPaidTo
						path
						upvotes
						meSats
						meDontLike
						meBookmark
						meSubscription
						outlawed
						freebie
						ncomments
						commentSats
						lastCommentAt
						maxBid
						isJob
						company
						location
						remote
						subName
						pollCost
						status
						uploadId
						mine
						position
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

	resp, err := MakeStackerNewsRequest(body)
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
	err = CheckForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}
	return &respBody.Data.Items, nil
}

// Create a new LINK post
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

// Create a new comment
func CreateComment(parentId int, text string) (int, error) {
	body := GraphQLPayload{
		Query: `
			mutation upsertComment($text: String!, $parentId: ID!) {
			  upsertComment(text: $text, parentId: $parentId) {
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
