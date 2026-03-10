package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/soda-data-inc/soda-cli/internal/config"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

type Client struct {
	baseURL      string
	apiKeyID     string
	apiKeySecret string
	http         *http.Client
}

func New(p config.Profile) *Client {
	host := p.Host
	if host == "" {
		host = "cloud.soda.io"
	}
	return &Client{
		baseURL:      "https://" + host,
		apiKeyID:     p.APIKeyID,
		apiKeySecret: p.APIKeySecret,
		http:         &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) post(path string, body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.apiKeyID, c.apiKeySecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return c.http.Do(req)
}

func (c *Client) patch(path string, body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PATCH", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.apiKeyID, c.apiKeySecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return c.http.Do(req)
}

func (c *Client) delete(path string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.apiKeyID, c.apiKeySecret)
	req.Header.Set("Accept", "application/json")
	return c.http.Do(req)
}

func (c *Client) get(path string, params url.Values) (*http.Response, error) {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.apiKeyID, c.apiKeySecret)
	req.Header.Set("Accept", "application/json")
	return c.http.Do(req)
}

func decode(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return &output.ExitError{Code: 3, Msg: "authentication failed — run `soda auth login`"}
	}
	if resp.StatusCode == 429 {
		return &output.ExitError{Code: 2, Msg: "rate limit exceeded — try again in a moment"}
	}
	if resp.StatusCode >= 400 {
		// try to extract a human-readable message from the JSON error body
		var apiErr struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return &output.ExitError{Code: 2, Msg: apiErr.Message}
		}
		return &output.ExitError{Code: 2, Msg: fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body))}
	}
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, target)
}
