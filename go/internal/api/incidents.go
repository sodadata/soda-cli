package api

// ── Incidents ─────────────────────────────────────────────────────────────────

type IncidentDataset struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Incident struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Severity    string          `json:"severity"`
	Status      string          `json:"status"`
	AssignedTo  string          `json:"assignedTo"`
	Dataset     IncidentDataset `json:"dataset"`
	CreatedAt   string          `json:"createdAt"`
	UpdatedAt   string          `json:"updatedAt"`
}

type UpdateIncidentRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Severity    *string `json:"severity,omitempty"`
	Status      *string `json:"status,omitempty"`
	AssignedTo  *string `json:"assignedTo,omitempty"`
}

func (c *Client) GetIncident(incidentID string) (*Incident, error) {
	resp, err := c.get("/api/v1/incidents/"+incidentID, nil)
	if err != nil {
		return nil, err
	}
	var result Incident
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateIncident(incidentID string, req UpdateIncidentRequest) (*Incident, error) {
	resp, err := c.patch("/api/v1/incidents/"+incidentID, req)
	if err != nil {
		return nil, err
	}
	var result Incident
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
