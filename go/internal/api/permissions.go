package api

// ── Dataset roles ──────────────────────────────────────────────────────────────

type DatasetRole struct {
	ID                      string `json:"id"`
	Name                    string `json:"name"`
	ViewProfilingAndSamples bool   `json:"viewProfilingAndSamples"`
	ViewFailedRows          bool   `json:"viewFailedRows"`
	ManageChecks            bool   `json:"manageChecks"`
	ConfigureDataset        bool   `json:"configureDataset"`
	ManagePermissions       bool   `json:"managePermissions"`
	ManageIncidents         bool   `json:"manageIncidents"`
	DeleteDataset           bool   `json:"deleteDataset"`
	PublishContracts        bool   `json:"publishContracts"`
	ManageContracts         bool   `json:"manageContracts"`
}

type DatasetRolesResponse struct {
	Content []DatasetRole `json:"content"`
}

func (c *Client) ListDatasetRoles() (*DatasetRolesResponse, error) {
	resp, err := c.get("/api/v1/datasets/roles", nil)
	if err != nil {
		return nil, err
	}
	var result DatasetRolesResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ── Dataset responsibilities ───────────────────────────────────────────────────

type DatasetResponsibility struct {
	Managed     bool        `json:"managed"`
	Role        DatasetRole `json:"role"`
	Type        string      `json:"type"` // "user" or "userGroup"
	UserID      string      `json:"userId"`
	UserGroupID string      `json:"userGroupId"`
}

type DatasetResponsibilitiesResponse struct {
	Content []DatasetResponsibility `json:"content"`
}

type ResponsibilityRequest struct {
	RoleID      string `json:"roleId"`
	Type        string `json:"type"`
	UserID      string `json:"userId,omitempty"`
	UserGroupID string `json:"userGroupId,omitempty"`
}

type UpdateResponsibilitiesRequest struct {
	Responsibilities []ResponsibilityRequest `json:"responsibilities"`
}

func (c *Client) GetDatasetResponsibilities(datasetID string) (*DatasetResponsibilitiesResponse, error) {
	resp, err := c.get("/api/v1/datasets/"+datasetID+"/responsibilities", nil)
	if err != nil {
		return nil, err
	}
	var result DatasetResponsibilitiesResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateDatasetResponsibilities(datasetID string, req UpdateResponsibilitiesRequest) (*DatasetResponsibilitiesResponse, error) {
	resp, err := c.post("/api/v1/datasets/"+datasetID+"/responsibilities", req)
	if err != nil {
		return nil, err
	}
	var result DatasetResponsibilitiesResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
