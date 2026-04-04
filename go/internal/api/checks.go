package api

import (
	"fmt"
	"net/url"
)

// ── Checks ────────────────────────────────────────────────────────────────────

type CheckDataset struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	QualifiedName string `json:"qualifiedName"`
}

type Check struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	EvaluationStatus string         `json:"evaluationStatus"`
	Column           string         `json:"column"`
	CheckType        string         `json:"checkType"`
	LastCheckRunTime string         `json:"lastCheckRunTime"`
	Datasets         []CheckDataset `json:"datasets"`
	Definition           string              `json:"definition"`
	Description          string              `json:"description"`
	CloudURL             string              `json:"cloudUrl"`
	CreatedAt            string              `json:"createdAt"`
	Attributes           map[string]any      `json:"attributes"`
	LastCheckResultValue *CheckResultValue   `json:"lastCheckResultValue"`
}

type CheckResultValue struct {
	Value      *float64 `json:"value"`
	ValueLabel string   `json:"valueLabel"`
}

type ChecksPage struct {
	Content       []Check `json:"content"`
	TotalElements int     `json:"totalElements"`
	Last          bool    `json:"last"`
}

type ListChecksParams struct {
	DatasetID string
	CheckIDs  string // comma-separated; cannot be combined with DatasetID
	Size      int
}

func (c *Client) ListChecks(p ListChecksParams) (*ChecksPage, error) {
	params := url.Values{}
	if p.CheckIDs != "" {
		params.Set("checkIds", p.CheckIDs)
	} else {
		size := p.Size
		if size < 10 {
			size = 10 // API enforces minimum of 10
		}
		params.Set("size", fmt.Sprintf("%d", size))
		if p.DatasetID != "" {
			params.Set("datasetId", p.DatasetID)
		}
	}
	resp, err := c.get("/api/v1/checks", params)
	if err != nil {
		return nil, err
	}
	var result ChecksPage
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
