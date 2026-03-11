package api

import (
	"fmt"
	"net/url"
)

type AgentVersions struct {
	Agent   string `json:"agent"`
	Library string `json:"library"`
}

type Agent struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Label             string        `json:"label"`
	Type              string        `json:"type"`
	IsOnline          bool          `json:"isOnline"`
	LastSeenTimestamp  string        `json:"lastSeenTimestamp"`
	Versions          AgentVersions `json:"versions"`
}

type AgentsPage struct {
	Content       []Agent `json:"content"`
	TotalElements int     `json:"totalElements"`
	Last          bool    `json:"last"`
}

type GetAgentResponse struct {
	Agent Agent `json:"agent"`
}

func (c *Client) ListAgents(size int) (*AgentsPage, error) {
	params := url.Values{}
	if size < 10 {
		size = 10
	}
	params.Set("size", fmt.Sprintf("%d", size))
	resp, err := c.get("/api/v1/agents", params)
	if err != nil {
		return nil, err
	}
	var result AgentsPage
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetAgent(agentID string) (*Agent, error) {
	resp, err := c.get("/api/v1/agents/"+agentID, nil)
	if err != nil {
		return nil, err
	}
	var result GetAgentResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result.Agent, nil
}
