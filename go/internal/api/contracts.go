package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

// ── Contracts ─────────────────────────────────────────────────────────────────

type Contract struct {
	ID                   string `json:"id"`
	DatasetID            string `json:"datasetId"`
	DatasetQualifiedName string `json:"datasetQualifiedName"`
	Contents             string `json:"contents"`
	Created              string `json:"created"`
	LastUpdated          string `json:"lastUpdated"`
}

type ContractPage struct {
	Content       []Contract `json:"content"`
	TotalElements int        `json:"totalElements"`
	TotalPages    int        `json:"totalPages"`
	Number        int        `json:"number"`
	Last          bool       `json:"last"`
}

type ContractRequest struct {
	DatasetID            string `json:"datasetId,omitempty"`
	DatasetQualifiedName string `json:"datasetQualifiedName,omitempty"`
	Contents             string `json:"contents"`
}

func (c *Client) ListContracts(page, size int) (*ContractPage, error) {
	params := url.Values{}
	if size <= 0 {
		size = 100
	}
	params.Set("size", strconv.Itoa(size))
	params.Set("page", strconv.Itoa(page))
	resp, err := c.get("/api/v1/contracts", params)
	if err != nil {
		return nil, err
	}
	var result ContractPage
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FindContractByDataset fetches pages until it finds a contract matching the
// given datasetQualifiedName, or exhausts all pages.
func (c *Client) FindContractByDataset(qualifiedName string) (*Contract, error) {
	page := 0
	for {
		p, err := c.ListContracts(page, 100)
		if err != nil {
			return nil, err
		}
		for i := range p.Content {
			if p.Content[i].DatasetQualifiedName == qualifiedName {
				return &p.Content[i], nil
			}
		}
		if p.Last {
			break
		}
		page++
	}
	return nil, nil // not found
}

func (c *Client) CreateContract(req ContractRequest) (*Contract, error) {
	resp, err := c.post("/api/v1/contracts", req)
	if err != nil {
		return nil, err
	}
	return decodeContractResponse(resp)
}

func (c *Client) UpdateContract(contractID string, req ContractRequest) (*Contract, error) {
	resp, err := c.post("/api/v1/contracts/"+contractID, req)
	if err != nil {
		return nil, err
	}
	return decodeContractResponse(resp)
}

func (c *Client) DeleteContract(contractID string) error {
	resp, err := c.delete("/api/v1/contracts/" + contractID)
	if err != nil {
		return err
	}
	var result struct{}
	return decode(resp, &result)
}

// decodeContractResponse handles two response shapes from the contract API:
//   - wrapped:  {"contract": {id, datasetId, ...}, "contents": "..."}
//   - flat:     {id, datasetId, datasetQualifiedName, ...}
func decodeContractResponse(resp *http.Response) (*Contract, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, &output.ExitError{Code: 3, Msg: "authentication failed — run `soda auth login`"}
	}
	if resp.StatusCode >= 400 {
		var apiErr struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return nil, &output.ExitError{Code: 2, Msg: apiErr.Message}
		}
		return nil, &output.ExitError{Code: 2, Msg: string(body)}
	}

	// Try wrapped format first: {"contract": {...}}
	var wrapper struct {
		Contract *Contract `json:"contract"`
	}
	if json.Unmarshal(body, &wrapper) == nil && wrapper.Contract != nil && wrapper.Contract.ID != "" {
		return wrapper.Contract, nil
	}

	// Fall back to flat format: {id, datasetId, ...}
	var flat Contract
	if err := json.Unmarshal(body, &flat); err != nil {
		return nil, err
	}
	return &flat, nil
}
