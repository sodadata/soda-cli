package api

// ── Scan status ──────────────────────────────────────────────────────────────

type ScanCheck struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"`   // dataset or column
	Column           string `json:"column"` // column name if type=column
	Definition       string `json:"definition"`
	EvaluationStatus string `json:"evaluationStatus"` // pass|warn|fail|notEvaluated|excluded
	Value            *float64 `json:"value"`
}

type ScanStatus struct {
	ID       string      `json:"id"`
	State    string      `json:"state"` // queuing|executing|started|completed|failed|...
	CloudURL string      `json:"cloudUrl"`
	Checks   []ScanCheck `json:"checks"`
	Failures int         `json:"failures"`
	Warnings int         `json:"warnings"`
	Errors   int         `json:"errors"`
	Started  string      `json:"started"`
	Ended    string      `json:"ended"`
	Created  string      `json:"created"`
}

// IsScanTerminal returns true if the scan state is a terminal state.
func IsScanTerminal(state string) bool {
	switch state {
	case "completed", "completedWithErrors", "completedWithFailures", "completedWithWarnings",
		"canceled", "timedOut", "failed":
		return true
	}
	return false
}

func (c *Client) GetScanStatus(scanID string) (*ScanStatus, error) {
	resp, err := c.get("/api/v1/scans/"+scanID, nil)
	if err != nil {
		return nil, err
	}
	var result ScanStatus
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CancelScan requests cancellation of a running scan.
func (c *Client) CancelScan(scanID string) error {
	resp, err := c.post("/api/v1/scans/"+scanID+"/actions/cancel", struct{}{})
	if err != nil {
		return err
	}
	return decode(resp, &struct{}{})
}

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
