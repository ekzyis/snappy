package sn

import (
	"encoding/json"
	"fmt"
)

func HasNewNotes() (bool, error) {
	body := GraphQLPayload{
		Query: `
			{
				hasNewNotes
			}`,
	}
	resp, err := MakeStackerNewsRequest(body)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var respBody HasNewNotesResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding SN hasNewNotes: %w", err)
		return false, err
	}
	err = CheckForErrors(respBody.Errors)
	if err != nil {
		return false, err
	}

	return respBody.Data.HasNewNotes, nil
}
