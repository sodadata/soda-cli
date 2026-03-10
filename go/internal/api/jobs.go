package api

// ── Scan logs ─────────────────────────────────────────────────────────────────

type ScanLogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type ScanLogsResponse struct {
	Content []ScanLogEntry `json:"content"`
	Logs    []ScanLogEntry `json:"logs"`
}

func (c *Client) GetScanLogs(scanID string) (*ScanLogsResponse, error) {
	resp, err := c.get("/api/v1/scans/"+scanID+"/logs", nil)
	if err != nil {
		return nil, err
	}
	var result ScanLogsResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
