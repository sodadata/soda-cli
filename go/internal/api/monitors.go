package api

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

type UpdateMetricMonitoringRequest struct {
	Enabled                            *bool                     `json:"enabled,omitempty"`
	ScanSchedule                       *ScanSchedule             `json:"scanSchedule,omitempty"`
	DatasetMetricMonitorsConfiguration []DatasetMetricMonitorCfg `json:"datasetMetricMonitorsConfiguration,omitempty"`
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

func (c *Client) UpdateMetricMonitoring(datasetID string, req UpdateMetricMonitoringRequest) (*MetricMonitoringConfig, error) {
	resp, err := c.post("/api/v1/datasets/"+datasetID+"/metricMonitoring", req)
	if err != nil {
		return nil, err
	}
	var result MetricMonitoringConfig
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
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
