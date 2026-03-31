package api

import (
	"net/url"
	"strconv"
)

// ── Incidents ─────────────────────────────────────────────────────────────────

type IncidentDataset struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Incident struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Number          int             `json:"number"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Severity        string          `json:"severity"`
	Status          string          `json:"status"`
	ResolutionNotes string          `json:"resolutionNotes"`
	CloudURL        string          `json:"cloudUrl"`
	AssignedTo      string          `json:"assignedTo"`
	Dataset         IncidentDataset `json:"dataset"`
	CreatedAt       string          `json:"createdAt"`
	UpdatedAt       string          `json:"updatedAt"`
}

// DisplayName returns the best available display name for the incident.
func (i *Incident) DisplayName() string {
	if i.Name != "" {
		return i.Name
	}
	return i.Title
}

type IncidentPage struct {
	Content       []Incident `json:"content"`
	TotalElements int        `json:"totalElements"`
	TotalPages    int        `json:"totalPages"`
	Number        int        `json:"number"`
	Last          bool       `json:"last"`
}

type UpdateIncidentRequest struct {
	Title       *string `json:"title,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Severity    *string `json:"severity,omitempty"`
	Status      *string `json:"status,omitempty"`
	AssignedTo  *string `json:"assignedTo,omitempty"`
}

func (c *Client) ListIncidents(page, size int) (*IncidentPage, error) {
	params := url.Values{}
	if size <= 0 {
		size = 100
	}
	params.Set("size", strconv.Itoa(size))
	params.Set("page", strconv.Itoa(page))
	resp, err := c.get("/api/v1/incidents", params)
	if err != nil {
		return nil, err
	}
	var result IncidentPage
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
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
	resp, err := c.post("/api/v1/incidents/"+incidentID, req)
	if err != nil {
		return nil, err
	}
	var result Incident
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
