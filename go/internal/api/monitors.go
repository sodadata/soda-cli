package api

import "time"

// ── Metric monitoring config ───────────────────────────────────────────────────

// ── Dataset metric monitors ───────────────────────────────────────────────────

type DatasetMonitorConfig struct {
	IsEnabled bool `json:"isEnabled"`
}

type DatasetMetricMonitorCfg struct {
	MetricType    string               `json:"metricType"`
	Configuration DatasetMonitorConfig `json:"configuration"`
}

// ── Column metric monitors ─────────────────────────────────────────────────────

type GroupByColumn struct {
	ColumnName     string   `json:"columnName"`
	ExcludedValues []string `json:"excludedValues,omitempty"`
}

type ColumnMonitorConfig struct {
	IsEnabled    bool            `json:"isEnabled"`
	GroupByColumns []GroupByColumn `json:"groupByColumns,omitempty"`
}

type ColumnMonitor struct {
	CheckID       string             `json:"checkId"`
	ColumnName    string             `json:"columnName"`
	MetricType    string             `json:"metricType"`
	Configuration ColumnMonitorConfig `json:"configuration"`
}

// ── Custom SQL monitors ───────────────────────────────────────────────────────

type CustomSqlMonitorConfig struct {
	SQLQuery     string `json:"sqlQuery"`
	ResultMetric string `json:"resultMetric"`
	IsEnabled    bool   `json:"isEnabled"`
}

type CustomSqlMonitor struct {
	CheckID       string                `json:"checkId"`
	MonitorName   string                `json:"monitorName"`
	ColumnName    string                `json:"columnName"`
	Configuration CustomSqlMonitorConfig `json:"configuration"`
}

// ── Metric monitoring config (GET response + POST request) ────────────────────

type MetricMonitoringConfig struct {
	DatasetID                              string                    `json:"datasetId"`
	Enabled                                bool                      `json:"enabled"`
	ScanSchedule                           *ScanSchedule             `json:"scanSchedule"`
	HistoricalMetricCollectionScanStartDate string                   `json:"historicalMetricCollectionScanStartDate"`
	DatasetMetricMonitorsConfiguration     []DatasetMetricMonitorCfg `json:"datasetMetricMonitorsConfiguration"`
	ColumnMetricMonitors                   []ColumnMonitor           `json:"columnMetricMonitors"`
	CustomSqlMetricMonitors                []CustomSqlMonitor        `json:"customSqlMetricMonitors"`
}

func (c *Client) GetMetricMonitoring(datasetID string) (*MetricMonitoringConfig, error) {
	resp, err := c.get("/api/v1/datasets/"+datasetID+"/metricMonitoring", nil)
	if err != nil {
		return nil, err
	}
	var result MetricMonitoringConfig
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateMetricMonitoring(datasetID string, req MetricMonitoringSettings) (*MetricMonitoringConfig, error) {
	// Use the dataset update endpoint (POST /api/v1/datasets/{id}) with the
	// metricMonitoring field — the dedicated /metricMonitoring sub-resource
	// is not available on all deployments.
	updateReq := UpdateDatasetRequest{MetricMonitoring: &req}
	resp, err := c.post("/api/v1/datasets/"+datasetID, updateReq)
	if err != nil {
		return nil, err
	}
	// Drain and close body — the POST returns a Dataset, not MetricMonitoringConfig.
	var discard Dataset
	if err := decode(resp, &discard); err != nil {
		return nil, err
	}
	// Re-fetch the monitoring config in the expected shape.
	return c.GetMetricMonitoring(datasetID)
}

// EnableDefaultMonitoring enables all dataset-level metric monitors for a dataset.
// It uses POST /api/v1/datasets/{id} (the generic dataset update endpoint) because
// defaultDatasetMonitorTypes are the known API metricType values for dataset-level monitors.
var defaultDatasetMonitorTypes = []string{
	"rowCount", "freshness", "schema", "rowsInserted", "totalRowCountChange", "timeliness",
}

// EnableDatasetDefaults enables metric monitoring and/or profiling for a dataset
// in a single POST /api/v1/datasets/{id} call.
func (c *Client) EnableDatasetDefaults(datasetID string, monitoring, profiling bool) error {
	if !monitoring && !profiling {
		return nil
	}

	req := UpdateDatasetRequest{}

	if profiling {
		t := true
		req.Profiling = &ProfilingSettings{
			Enabled: &t,
			ProfilingSamplingStrategy: &SamplingStrategy{
				NumberOfRows: 1000000,
			},
		}
	}

	if monitoring {
		t := true
		monitors := make([]DatasetMetricMonitorCfg, len(defaultDatasetMonitorTypes))
		for i, mt := range defaultDatasetMonitorTypes {
			monitors[i] = DatasetMetricMonitorCfg{
				MetricType:    mt,
				Configuration: DatasetMonitorConfig{IsEnabled: true},
			}
		}
		req.MetricMonitoring = &MetricMonitoringSettings{
			Enabled: &t,
			ScanSchedule: &ScanSchedule{
				CronExpression: "0 6 * * *",
				Timezone:       "UTC",
			},
			DatasetMetricMonitorsConfiguration:      monitors,
			HistoricalMetricCollectionScanStartDate: time.Now().AddDate(0, 0, -30).UTC().Format(time.RFC3339),
		}
	}

	_, err := c.UpdateDataset(datasetID, req)
	return err
}

// ── Column metric monitors ─────────────────────────────────────────────────────

type ColumnMetricMonitorCfg struct {
	MetricType    string             `json:"metricType"`
	Configuration ColumnMonitorConfig `json:"configuration"`
}

type CreateColumnMonitorRequest struct {
	ColumnName                       string                 `json:"columnName"`
	ColumnMetricMonitorConfiguration ColumnMetricMonitorCfg `json:"columnMetricMonitorConfiguration"`
}

type DeleteMonitorResponse struct {
	Message string `json:"message"`
}

func (c *Client) CreateColumnMonitor(datasetID string, req CreateColumnMonitorRequest) (*ColumnMonitor, error) {
	resp, err := c.post("/api/v1/datasets/"+datasetID+"/columnMetricMonitors", req)
	if err != nil {
		return nil, err
	}
	var result ColumnMonitor
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type UpdateColumnMonitorRequest struct {
	Configuration ColumnMonitorConfig `json:"configuration"`
}

func (c *Client) UpdateColumnMonitor(datasetID, monitorID string, req UpdateColumnMonitorRequest) (*ColumnMonitor, error) {
	resp, err := c.post("/api/v1/datasets/"+datasetID+"/columnMetricMonitors/"+monitorID, req)
	if err != nil {
		return nil, err
	}
	var result ColumnMonitor
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteColumnMonitor(datasetID, monitorID string) (*DeleteMonitorResponse, error) {
	resp, err := c.delete("/api/v1/datasets/" + datasetID + "/columnMetricMonitors/" + monitorID)
	if err != nil {
		return nil, err
	}
	var result DeleteMonitorResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Custom SQL monitors ────────────────────────────────────────────────────────

type CreateCustomSqlMonitorRequest struct {
	MonitorName   string               `json:"monitorName"`
	Configuration CustomSqlMonitorConfig `json:"configuration"`
	ColumnName    string               `json:"columnName,omitempty"`
}

func (c *Client) CreateCustomSqlMonitor(datasetID string, req CreateCustomSqlMonitorRequest) (*CustomSqlMonitor, error) {
	resp, err := c.post("/api/v1/datasets/"+datasetID+"/customSqlMonitors", req)
	if err != nil {
		return nil, err
	}
	var result CustomSqlMonitor
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type UpdateCustomSqlMonitorRequest struct {
	MonitorName   string                `json:"monitorName,omitempty"`
	Configuration CustomSqlMonitorConfig `json:"configuration"`
	ColumnName    string                `json:"columnName,omitempty"`
}

func (c *Client) UpdateCustomSqlMonitor(datasetID, monitorID string, req UpdateCustomSqlMonitorRequest) (*CustomSqlMonitor, error) {
	resp, err := c.post("/api/v1/datasets/"+datasetID+"/customSqlMonitors/"+monitorID, req)
	if err != nil {
		return nil, err
	}
	var result CustomSqlMonitor
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteCustomSqlMonitor(datasetID, monitorID string) (*DeleteMonitorResponse, error) {
	resp, err := c.delete("/api/v1/datasets/" + datasetID + "/customSqlMonitors/" + monitorID)
	if err != nil {
		return nil, err
	}
	var result DeleteMonitorResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
