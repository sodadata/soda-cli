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
