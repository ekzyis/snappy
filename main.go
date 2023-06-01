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
	SnUrl    = "https://stacker.news"
	SnApiUrl = "https://stacker.news/api/graphql"
	// TODO add API key support
	// SnApiKey string
	SnAuthCookie string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	flag.StringVar(&SnAuthCookie, "SN_AUTH_COOKIE", "", "Cookie required for authorizing requests to stacker.news/api/graphql")
	flag.Parse()
	if SnAuthCookie == "" {
		log.Fatal("SN_AUTH_COOKIE not set")
	}
}

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

func FormatLink(id int) string {
	return fmt.Sprintf("%s/items/%d", SnUrl, id)
}
