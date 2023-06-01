// Package for stacker.news API access
package sn

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/namsral/flag"
)

var (
	// stacker.news URL
	SnUrl = "https://stacker.news"
	// stacker.news API URL
	SnApiUrl = "https://stacker.news/api/graphql"
	// stacker.news session cookie
	SnAuthCookie string
	// TODO add API key support
	// SnApiKey string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	flag.StringVar(&SnAuthCookie, "SN_AUTH_COOKIE", "", "Cookie required for authentication requests to stacker.news/api/graphql")
	flag.Parse()
	if SnAuthCookie == "" {
		log.Fatal("SN_AUTH_COOKIE not set")
	}
}

// Make GraphQL request using raw payload
func MakeStackerNewsRequest(body GraphQLPayload) (*http.Response, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		err = fmt.Errorf("error encoding SN payload: %w", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", SnApiUrl, bytes.NewBuffer(bodyJSON))
	if err != nil {
		err = fmt.Errorf("error preparing SN request: %w", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Cookie", SnAuthCookie)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("error posting SN payload: %w", err)
		return nil, err
	}

	return resp, nil
}

// Returns error if any error was found
func CheckForErrors(graphqlErrors []GraphQLError) error {
	if len(graphqlErrors) > 0 {
		errorMsg, marshalErr := json.Marshal(graphqlErrors)
		if marshalErr != nil {
			return marshalErr
		}
		return errors.New(string(errorMsg))
	}
	return nil
}

// Format item id as link
func FormatLink(id int) string {
	return fmt.Sprintf("%s/items/%d", SnUrl, id)
}
