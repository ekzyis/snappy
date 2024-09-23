package sn

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

type Client struct {
	BaseUrl  string
	ApiUrl   string
	ApiKey   string
	MediaUrl string
}

func NewClient(options ...func(*Client)) *Client {
	c := &Client{}
	for _, o := range options {
		o(c)
	}

	// set defaults
	if c.BaseUrl == "" {
		c.BaseUrl = "https://stacker.news"
	}
	if c.ApiKey == "" {
		c.ApiKey = os.Getenv("SN_API_KEY")
	}
	if c.MediaUrl == "" {
		c.MediaUrl = "https://m.stacker.news"
	}
	c.ApiUrl = fmt.Sprintf("%s/api/graphql", c.BaseUrl)

	return c
}

func WithApiKey(apiKey string) func(*Client) {
	return func(c *Client) {
		c.ApiKey = apiKey
	}
}

func WithBaseUrl(baseUrl string) func(*Client) {
	return func(c *Client) {
		c.BaseUrl = baseUrl
	}
}

func WithMediaUrl(mediaUrl string) func(*Client) {
	return func(c *Client) {
		c.MediaUrl = mediaUrl
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
