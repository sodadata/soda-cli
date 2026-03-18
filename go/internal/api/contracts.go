package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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

// ── Skeleton / AI generation (async) ─────────────────────────────────────────

type CreateSkeletonRequest struct {
	DatasetID            string `json:"datasetId,omitempty"`
	DatasetQualifiedName string `json:"datasetQualifiedName,omitempty"`
}

type SkeletonStatus struct {
	OperationID          string `json:"operationId"`
	State                string `json:"state"` // ongoing, completed, failed, canceled
	DatasetID            string `json:"datasetId"`
	DatasetQualifiedName string `json:"datasetQualifiedName"`
	Created              string `json:"created"`
}

type GenerateContractRequest struct {
	DatasetIDs            []string `json:"datasetIds,omitempty"`
	DatasetQualifiedNames []string `json:"datasetQualifiedNames,omitempty"`
}

type GenerateDatasetStatus struct {
	DatasetID            string `json:"datasetId"`
	DatasetQualifiedName string `json:"datasetQualifiedName"`
	ScanID               string `json:"scanId"`
	ScanCloudURL         string `json:"scanCloudUrl"`
	ScanState            string `json:"scanState"`
}

type GenerateStatus struct {
	OperationID string                  `json:"operationId"`
	State       string                  `json:"state"` // ongoing, completed, failed, canceled
	Created     string                  `json:"created"`
	Datasets    []GenerateDatasetStatus `json:"datasets"`
}

func (c *Client) CreateSkeleton(req CreateSkeletonRequest) (string, error) {
	resp, err := c.post("/api/v1/contracts/actions/createSkeleton", req)
	if err != nil {
		return "", err
	}
	return decodeAsyncOperation(resp)
}

func (c *Client) GetSkeletonStatus(operationID string) (*SkeletonStatus, error) {
	resp, err := c.get("/api/v1/contracts/actions/createSkeleton/"+operationID, nil)
	if err != nil {
		return nil, err
	}
	var result SkeletonStatus
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GenerateContract(req GenerateContractRequest) (string, error) {
	resp, err := c.post("/api/v1/contracts/actions/generate", req)
	if err != nil {
		return "", err
	}
	return decodeAsyncOperation(resp)
}

// decodeAsyncOperation handles async API responses that return 202 + Location header.
// On error it returns the API's message field rather than a generic auth error.
func decodeAsyncOperation(resp *http.Response) (string, error) {
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		var apiErr struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			code := 2
			if resp.StatusCode == 401 {
				code = 3
			}
			return "", &output.ExitError{Code: code, Msg: apiErr.Message}
		}
		if resp.StatusCode == 401 {
			return "", &output.ExitError{Code: 3, Msg: "authentication failed — run `soda auth login`"}
		}
		return "", &output.ExitError{Code: 2, Msg: fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body))}
	}
	return extractOperationID(resp.Header.Get("Location")), nil
}

// extractOperationID gets the last path segment from a Location URL.
func extractOperationID(location string) string {
	if location == "" {
		return ""
	}
	parts := strings.Split(strings.TrimRight(location, "/"), "/")
	return parts[len(parts)-1]
}

func (c *Client) GetGenerateStatus(operationID string) (*GenerateStatus, error) {
	resp, err := c.get("/api/v1/contracts/actions/generate/"+operationID, nil)
	if err != nil {
		return nil, err
	}
	var result GenerateStatus
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// VerifyContract triggers a contract verification via Soda Cloud.
// Returns the scan ID from the X-Soda-Scan-Id response header.
func (c *Client) VerifyContract(contractID string) (string, error) {
	resp, err := c.post("/api/v1/contracts/"+contractID+"/verify", struct{}{})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return "", &output.ExitError{Code: 3, Msg: "authentication failed — run `soda auth login`"}
	}
	if resp.StatusCode >= 400 {
		var apiErr struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return "", &output.ExitError{Code: 2, Msg: apiErr.Message}
		}
		return "", &output.ExitError{Code: 2, Msg: fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body))}
	}

	// Extract scan ID from header or Location
	scanID := resp.Header.Get("X-Soda-Scan-Id")
	if scanID == "" {
		scanID = extractOperationID(resp.Header.Get("Location"))
	}
	if scanID == "" {
		// Try to extract from response body
		var result struct {
			ScanID string `json:"scanId"`
			ID     string `json:"id"`
		}
		if json.Unmarshal(body, &result) == nil {
			if result.ScanID != "" {
				scanID = result.ScanID
			} else if result.ID != "" {
				scanID = result.ID
			}
		}
	}
	if scanID == "" {
		return "", &output.ExitError{Code: 2, Msg: "verify request succeeded but no scan ID was returned"}
	}
	return scanID, nil
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
