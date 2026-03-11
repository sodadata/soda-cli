package api

type Datasource struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
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
