package sn

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

type ClientOptions struct {
	BaseUrl string
	ApiKey  string
}

type Client struct {
	BaseUrl string
	ApiUrl  string
	ApiKey  string
}

func NewClient(options *ClientOptions) *Client {
	if options.BaseUrl == "" {
		options.BaseUrl = "https://stacker.news"
	}
	if options.ApiKey == "" {
		options.ApiKey = os.Getenv("SN_API_KEY")
	}

	return &Client{
		BaseUrl: options.BaseUrl,
		ApiUrl:  fmt.Sprintf("%s/api/graphql", options.BaseUrl),
		ApiKey:  options.ApiKey,
	}
}

type GqlBody struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GqlError struct {
	Message string `json:"message"`
}

func (c *Client) callApi(body GqlBody) (*http.Response, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		err = fmt.Errorf("error encoding SN payload: %w", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", c.ApiUrl, bytes.NewBuffer(bodyJSON))
	if err != nil {
		err = fmt.Errorf("error preparing SN request: %w", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.ApiKey != "" {
		req.Header.Set("X-Api-Key", c.ApiKey)
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) checkForErrors(err []GqlError) error {
	if len(err) > 0 {
		errMsg, marshalErr := json.Marshal(err)
		if marshalErr != nil {
			return marshalErr
		}
		return errors.New(string(errMsg))
	}
	return nil
}
