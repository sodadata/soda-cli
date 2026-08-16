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
	RunnerID                  string `json:"runnerId,omitempty"`
	ConfigurationFileContents string `json:"configurationFileContents,omitempty"`
	Label                     string `json:"label,omitempty"`
}

func (c *Client) UpdateDatasource(datasourceID string, req UpdateDatasourceRequest) (*Datasource, error) {
	resp, err := c.post("/api/v1/datasources/"+datasourceID, req)
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
		return nil, &output.ExitError{Code: 3, Msg: "authentication failed — run `sodacli auth login`"}
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
	RunnerID                  string `json:"runnerId"`
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
	RunnerID                  string `json:"runnerId"`
	ConfigurationFileContents string `json:"configurationFileContents"`
}

type TestConnectionResponse struct {
	OperationID string `json:"operationId"`
}

type TestConnectionStatus struct {
	ID      string `json:"id"`
	State   string `json:"state"` // queued, processing, completed, failed, cancelled
	Message string `json:"message"`
	Started string `json:"started"`
}

func (c *Client) TestConnection(req TestConnectionRequest) (string, error) {
	resp, err := c.post("/api/v1/datasources/actions/testConnection", req)
	if err != nil {
		return "", err
	}
	return decodeAsyncTestConnection(resp)
}

// decodeAsyncTestConnection handles the 202 + Location response, falling back
// to a JSON body with operationId if the header is absent.
func decodeAsyncTestConnection(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return "", &output.ExitError{Code: 3, Msg: "authentication failed — run `sodacli auth login`"}
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
	// Try Location header first
	if loc := resp.Header.Get("Location"); loc != "" {
		parts := strings.Split(strings.TrimRight(loc, "/"), "/")
		return parts[len(parts)-1], nil
	}
	// Fall back to JSON body
	var result TestConnectionResponse
	if len(body) > 0 {
		if err := json.Unmarshal(body, &result); err == nil && result.OperationID != "" {
			return result.OperationID, nil
		}
	}
	return "", &output.ExitError{Code: 2, Msg: "test-connection request succeeded but no operation ID was returned"}
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

// ── Datasource diagnostics warehouse ─────────────────────────────────────────

type DatasourceDiagnosticsActionButton struct {
	Enabled bool   `json:"enabled"`
	Title   string `json:"title"`
	URL     string `json:"url"`
}

type FailedRowsCollectionStrategy struct {
	Type               string  `json:"type"`               // useDefaultMaxRowCount, absolute, percentage
	Threshold          float64 `json:"threshold"`
	ThresholdCondition string  `json:"thresholdCondition"` // greaterThan, lessThan
	MaxRowCountOverride int    `json:"maxRowCountOverride,omitempty"`
}

type DatasourceDiagnosticsFailedRows struct {
	Enabled                      bool                              `json:"enabled"`
	ExposeQueries                bool                              `json:"exposeQueries"`
	MaxRowCount                  int                               `json:"maxRowCount"`
	FailedRowsCollectionStrategy *FailedRowsCollectionStrategy     `json:"failedRowsCollectionStrategy,omitempty"`
	LocationMessage              string                            `json:"locationMessage"`
	ActionButton                 DatasourceDiagnosticsActionButton `json:"actionButton"`
}

type DatasourceDiagnosticsScanResults struct {
	Enabled bool `json:"enabled"`
}

type DatasourceDiagnosticsWarehouse struct {
	Enabled                     bool                              `json:"enabled"`
	ReuseDatasource             bool                              `json:"reuseDatasource"`
	TableNameTemplate           string                            `json:"tableNameTemplate"`
	FailedRowsConfiguration     *DatasourceDiagnosticsFailedRows  `json:"failedRowsConfiguration"`
	ScanAndResultsConfiguration *DatasourceDiagnosticsScanResults `json:"scanAndResultsConfiguration"`
}

func (c *Client) GetDatasourceDiagnostics(datasourceID string) (*DatasourceDiagnosticsWarehouse, error) {
	resp, err := c.get("/api/v1/datasources/"+datasourceID+"/diagnosticsWarehouse", nil)
	if err != nil {
		return nil, err
	}
	var result DatasourceDiagnosticsWarehouse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update request types — use pointers + omitempty to send only changed fields.

type UpdateActionButton struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Title   string `json:"title,omitempty"`
	URL     string `json:"url,omitempty"`
}

type UpdateFailedRowsConfig struct {
	Enabled                      *bool                         `json:"enabled,omitempty"`
	ExposeQueries                *bool                         `json:"exposeQueries,omitempty"`
	MaxRowCount                  *int                          `json:"maxRowCount,omitempty"`
	LocationMessage              string                        `json:"locationMessage,omitempty"`
	ActionButton                 *UpdateActionButton           `json:"actionButton,omitempty"`
	FailedRowsCollectionStrategy *FailedRowsCollectionStrategy `json:"failedRowsCollectionStrategy,omitempty"`
}

type UpdateScanResultsConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type UpdateDatasourceDiagnosticsRequest struct {
	Enabled                     *bool                    `json:"enabled,omitempty"`
	ReuseDatasource             *bool                    `json:"reuseDatasource,omitempty"`
	ConfigurationFileContents   string                   `json:"configurationFileContents,omitempty"`
	TableNameTemplate           string                   `json:"tableNameTemplate,omitempty"`
	FailedRowsConfiguration     *UpdateFailedRowsConfig  `json:"failedRowsConfiguration,omitempty"`
	ScanAndResultsConfiguration *UpdateScanResultsConfig `json:"scanAndResultsConfiguration,omitempty"`
}

func (c *Client) UpdateDatasourceDiagnostics(datasourceID string, req UpdateDatasourceDiagnosticsRequest) (*DatasourceDiagnosticsWarehouse, error) {
	resp, err := c.post("/api/v1/datasources/"+datasourceID+"/diagnosticsWarehouse", req)
	if err != nil {
		return nil, err
	}
	var result DatasourceDiagnosticsWarehouse
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
