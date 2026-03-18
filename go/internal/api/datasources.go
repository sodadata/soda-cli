package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

type Datasource struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Label     string `json:"label"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ── List datasources ──────────────────────────────────────────────────────────

type DatasourcePage struct {
	Content       []Datasource `json:"content"`
	TotalElements int          `json:"totalElements"`
	TotalPages    int          `json:"totalPages"`
	Number        int          `json:"number"`
	Last          bool         `json:"last"`
}

func (c *Client) ListDatasources(page, size int) (*DatasourcePage, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	if size <= 0 {
		size = 100
	}
	params.Set("size", strconv.Itoa(size))
	resp, err := c.get("/api/v1/datasources", params)
	if err != nil {
		return nil, err
	}
	var result DatasourcePage
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetDatasource(datasourceID string) (*Datasource, error) {
	resp, err := c.get("/api/v1/datasources/"+datasourceID, nil)
	if err != nil {
		return nil, err
	}
	return decodeDatasourceResponse(resp)
}

type UpdateDatasourceRequest struct {
	AgentID                   string `json:"agentId,omitempty"`
	ConfigurationFileContents string `json:"configurationFileContents,omitempty"`
	Label                     string `json:"label,omitempty"`
}

func (c *Client) UpdateDatasource(datasourceID string, req UpdateDatasourceRequest) (*Datasource, error) {
	resp, err := c.patch("/api/v1/datasources/"+datasourceID, req)
	if err != nil {
		return nil, err
	}
	return decodeDatasourceResponse(resp)
}

// decodeDatasourceResponse handles {"datasource": {...}} or flat {...} shapes.
func decodeDatasourceResponse(resp *http.Response) (*Datasource, error) {
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
		return nil, &output.ExitError{Code: 2, Msg: fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body))}
	}

	var wrapper struct {
		Datasource *Datasource `json:"datasource"`
	}
	if json.Unmarshal(body, &wrapper) == nil && wrapper.Datasource != nil && wrapper.Datasource.ID != "" {
		return wrapper.Datasource, nil
	}
	var flat Datasource
	if err := json.Unmarshal(body, &flat); err != nil {
		return nil, err
	}
	return &flat, nil
}

// ── Discovered datasets ───────────────────────────────────────────────────────

type DiscoveredDataset struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	QualifiedName string `json:"qualifiedName"`
	DatasourceID  string `json:"datasourceId"`
	Onboarded     bool   `json:"onboarded"`
	CreatedAt     string `json:"createdAt"`
}

type DiscoveredDatasetPage struct {
	Content       []DiscoveredDataset `json:"content"`
	TotalElements int                 `json:"totalElements"`
	TotalPages    int                 `json:"totalPages"`
	Number        int                 `json:"number"`
	Last          bool                `json:"last"`
}

func (c *Client) ListDiscoveredDatasets(datasourceID string, page, size int) (*DiscoveredDatasetPage, error) {
	params := url.Values{}
	params.Set("datasourceId", datasourceID)
	params.Set("page", strconv.Itoa(page))
	if size <= 0 {
		size = 100
	}
	params.Set("size", strconv.Itoa(size))
	resp, err := c.get("/api/v1/discoveredDatasets", params)
	if err != nil {
		return nil, err
	}
	var result DiscoveredDatasetPage
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Onboard discovered datasets ───────────────────────────────────────────────

type OnboardDatasetsRequest struct {
	DiscoveredDatasetIDs []string `json:"discoveredDatasetIds"`
}

// OnboardDiscoveredDatasets onboards discovered datasets into Soda Cloud.
// Returns nil on success (the API returns 202 with no body).
func (c *Client) OnboardDiscoveredDatasets(datasourceID string, req OnboardDatasetsRequest) error {
	resp, err := c.post("/api/v1/datasources/"+datasourceID+"/onboardDatasets", req)
	if err != nil {
		return err
	}
	// 202 Accepted with empty body — just check for errors
	return decode(resp, &struct{}{})
}

type CreateDatasourceRequest struct {
	Name                      string `json:"name"`
	AgentID                   string `json:"agentId"`
	ConfigurationFileContents string `json:"configurationFileContents"`
	Label                     string `json:"label,omitempty"`
}

type CreateDatasourceResponse struct {
	Datasource      Datasource `json:"datasource"`
	DiscoveryScanID string     `json:"discoveryScanId"`
}

type DeleteDatasourceResponse struct {
	Message string `json:"message"`
}

func (c *Client) CreateDatasource(req CreateDatasourceRequest) (*CreateDatasourceResponse, error) {
	resp, err := c.post("/api/v1/datasources", req)
	if err != nil {
		return nil, err
	}
	var result CreateDatasourceResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Test connection (async) ──────────────────────────────────────────────────

type TestConnectionRequest struct {
	AgentID                   string `json:"agentId"`
	ConfigurationFileContents string `json:"configurationFileContents"`
}

type TestConnectionResponse struct {
	OperationID string `json:"operationId"`
}

type TestConnectionStatus struct {
	OperationID string `json:"operationId"`
	State       string `json:"state"` // "running", "succeeded", "failed"
	Message     string `json:"message"`
}

func (c *Client) TestConnection(req TestConnectionRequest) (*TestConnectionResponse, error) {
	resp, err := c.post("/api/v1/datasources/actions/testConnection", req)
	if err != nil {
		return nil, err
	}
	var result TestConnectionResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetTestConnectionStatus(operationID string) (*TestConnectionStatus, error) {
	resp, err := c.get("/api/v1/datasources/actions/testConnection/"+operationID, nil)
	if err != nil {
		return nil, err
	}
	var result TestConnectionStatus
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteDatasource(datasourceID string) (*DeleteDatasourceResponse, error) {
	resp, err := c.delete("/api/v1/datasources/" + datasourceID)
	if err != nil {
		return nil, err
	}
	var result DeleteDatasourceResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
