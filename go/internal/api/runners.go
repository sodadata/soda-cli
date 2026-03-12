package api

import (
	"fmt"
	"net/url"
)

type RunnerVersions struct {
	Agent   string `json:"agent"`
	Library string `json:"library"`
}

type Runner struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	Label             string         `json:"label"`
	Type              string         `json:"type"`
	IsOnline          bool           `json:"isOnline"`
	LastSeenTimestamp  string         `json:"lastSeenTimestamp"`
	Versions          RunnerVersions `json:"versions"`
}

type RunnersPage struct {
	Content       []Runner `json:"content"`
	TotalElements int      `json:"totalElements"`
	Last          bool     `json:"last"`
}

type GetRunnerResponse struct {
	Runner Runner `json:"runner"`
}

func (c *Client) ListRunners(size int) (*RunnersPage, error) {
	params := url.Values{}
	if size < 10 {
		size = 10
	}
	params.Set("size", fmt.Sprintf("%d", size))
	resp, err := c.get("/api/v1/runners", params)
	if err != nil {
		return nil, err
	}
	var result RunnersPage
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetRunner(runnerID string) (*Runner, error) {
	resp, err := c.get("/api/v1/runners/"+runnerID, nil)
	if err != nil {
		return nil, err
	}
	var result GetRunnerResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result.Runner, nil
}
