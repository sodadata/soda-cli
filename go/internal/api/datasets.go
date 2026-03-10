package api

import (
	"net/url"
	"strconv"
)

type DatasourceProperties struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Dataset struct {
	ID                string               `json:"id"`
	Name              string               `json:"name"`
	QualifiedName     string               `json:"qualifiedName"`
	CloudURL          string               `json:"cloudUrl"`
	Checks            float64              `json:"checks"`
	Incidents         float64              `json:"incidents"`
	DataQualityStatus string               `json:"dataQualityStatus"`
	Datasource        DatasourceProperties `json:"datasource"`
	Tags              []string             `json:"tags"`
	LastUpdated       string               `json:"lastUpdated"`
}

type DatasetPage struct {
	Content       []Dataset `json:"content"`
	TotalElements int       `json:"totalElements"`
	TotalPages    int       `json:"totalPages"`
	Number        int       `json:"number"`
	Last          bool      `json:"last"`
}

// ── Profiling ─────────────────────────────────────────────────────────────────

type ScanSchedule struct {
	CronExpression string `json:"cronExpression"`
	Timezone       string `json:"timezone"`
}

type SamplingStrategy struct {
	NumberOfRows  int    `json:"numberOfRows"`
	NumberOfUnits int    `json:"numberOfUnits"`
	UnitOfTime    string `json:"unitOfTime"`
}

type ColumnMetrics struct {
	Average           *float64 `json:"average"`
	Sum               *float64 `json:"sum"`
	Median            *float64 `json:"median"`
	Minimum           *float64 `json:"minimum"`
	Maximum           *float64 `json:"maximum"`
	MinimumLength     *int     `json:"minimumLength"`
	MaximumLength     *int     `json:"maximumLength"`
	AverageLength     *float64 `json:"averageLength"`
	MissingCount      *float64 `json:"missingCount"`
	DistinctCount     *float64 `json:"distinctCount"`
	StandardDeviation *float64 `json:"standardDeviation"`
}

type ProfilingColumn struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Type    string        `json:"type"`
	Metrics ColumnMetrics `json:"metrics"`
}

type ProfilingResult struct {
	DatasetID                string            `json:"datasetId"`
	Enabled                  bool              `json:"enabled"`
	RowCount                 *float64          `json:"rowCount"`
	ColumnCount              *int              `json:"columnCount"`
	ProfilingTime            string            `json:"profilingTime"`
	ScanSchedule             *ScanSchedule     `json:"scanSchedule"`
	SamplingStrategyConfig   *SamplingStrategy `json:"samplingStrategyConfiguration"`
	Columns                  []ProfilingColumn `json:"columns"`
}

func (c *Client) GetProfiling(datasetID string) (*ProfilingResult, error) {
	resp, err := c.get("/api/v1/datasets/"+datasetID+"/profiling", nil)
	if err != nil {
		return nil, err
	}
	var result ProfilingResult
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ProfilingSettings maps to the `profiling` field in POST /api/v1/datasets/{id}
type ProfilingSettings struct {
	Enabled                  *bool             `json:"enabled,omitempty"`
	ScanSchedule             *ScanSchedule     `json:"scanSchedule,omitempty"`
	ProfilingSamplingStrategy *SamplingStrategy `json:"profilingSamplingStrategy,omitempty"`
}

// ── Owner ─────────────────────────────────────────────────────────────────────

type DatasetOwnerRequest struct {
	Type        string `json:"type"` // "user" or "userGroup"
	UserID      string `json:"userId,omitempty"`
	UserGroupID string `json:"userGroupId,omitempty"`
}

// ── Diagnostics warehouse ─────────────────────────────────────────────────────

// GET /api/v1/datasets/{id}/diagnosticsWarehouse response
type DiagnosticsWarehouseResult struct {
	FailedRowsConfiguration     *DiagnosticsFailedRowsResult `json:"failedRowsConfiguration"`
	ScanAndResultsConfiguration *DiagnosticsScanResult       `json:"scanAndResultsConfiguration"`
}

type DiagnosticsFailedRowsResult struct {
	Enabled     bool   `json:"enabled"`
	MaxRowCount int    `json:"maxRowCount"`
	State       string `json:"state"`
}

type DiagnosticsScanResult struct {
	Enabled bool `json:"enabled"`
}

func (c *Client) GetDatasetDiagnostics(datasetID string) (*DiagnosticsWarehouseResult, error) {
	resp, err := c.get("/api/v1/datasets/"+datasetID+"/diagnosticsWarehouse", nil)
	if err != nil {
		return nil, err
	}
	var result DiagnosticsWarehouseResult
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// POST /api/v1/datasets/{id}/diagnosticsWarehouse request
type DiagnosticsFailedRowsConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type DiagnosticsScanConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type DiagnosticsWarehouseConfig struct {
	FailedRowsConfiguration     *DiagnosticsFailedRowsConfig `json:"failedRowsConfiguration,omitempty"`
	ScanAndResultsConfiguration *DiagnosticsScanConfig       `json:"scanAndResultsConfiguration,omitempty"`
}

func (c *Client) UpdateDatasetDiagnostics(datasetID string, cfg DiagnosticsWarehouseConfig) (*Dataset, error) {
	return c.UpdateDataset(datasetID, UpdateDatasetRequest{DiagnosticsWarehouse: &cfg})
}

// ── Time partition ────────────────────────────────────────────────────────────

type TimePartitionRequest struct {
	PartitionColumn string `json:"partitionColumn,omitempty"`
}

// ── Update request ────────────────────────────────────────────────────────────

type UpdateDatasetRequest struct {
	Profiling            *ProfilingSettings          `json:"profiling,omitempty"`
	Owners               []DatasetOwnerRequest       `json:"owners,omitempty"`
	Tags                 []string                    `json:"tags,omitempty"`
	TimePartition        *TimePartitionRequest       `json:"timePartition,omitempty"`
	DiagnosticsWarehouse *DiagnosticsWarehouseConfig `json:"diagnosticsWarehouse,omitempty"`
}

func (c *Client) UpdateDataset(datasetID string, req UpdateDatasetRequest) (*Dataset, error) {
	resp, err := c.post("/api/v1/datasets/"+datasetID, req)
	if err != nil {
		return nil, err
	}
	var result Dataset
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateDatasetProfiling(datasetID string, settings ProfilingSettings) (*Dataset, error) {
	return c.UpdateDataset(datasetID, UpdateDatasetRequest{Profiling: &settings})
}

// ── Delete ────────────────────────────────────────────────────────────────────

type DeleteDatasetResponse struct {
	Message string `json:"message"`
}

func (c *Client) DeleteDataset(datasetID string) (*DeleteDatasetResponse, error) {
	resp, err := c.delete("/api/v1/datasets/" + datasetID)
	if err != nil {
		return nil, err
	}
	var result DeleteDatasetResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type ListDatasetsParams struct {
	DatasourceName string
	Search         string
	Page           int
	Size           int
}

func (c *Client) ListDatasets(p ListDatasetsParams) (*DatasetPage, error) {
	params := url.Values{}
	if p.DatasourceName != "" {
		params.Set("datasourceName", p.DatasourceName)
	}
	if p.Search != "" {
		params.Set("search", p.Search)
	}
	size := p.Size
	if size <= 0 {
		size = 100
	}
	params.Set("size", strconv.Itoa(size))
	params.Set("page", strconv.Itoa(p.Page))

	resp, err := c.get("/api/v1/datasets", params)
	if err != nil {
		return nil, err
	}
	var result DatasetPage
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
