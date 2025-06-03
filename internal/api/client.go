package api

import (
	"github.com/go-resty/resty/v2"
	"github.com/piersCraft/llm-entity-resolution/internal/domain"
)

type APIClient struct {
	client *resty.Client
	config *config.Config
}

func NewClient(cfg *config.Config) *APIClient {
	return &APIClient{
		client: resty.New().SetTimeout(30 * time.Second),
		config: cfg,
	}
}

func (c *APIClient) ProcessData(data string) (*domain.OutputRecord, error) {
	resp, err := c.client.R().
		SetAuthToken(c.config.APIToken).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"data": data}).
		Post(c.config.APIEndpoint)
	if err != nil {
		return nil, err
	}

	return &domain.OutputRecord{
		InputData:  data,
		Response:   resp.String(),
		StatusCode: resp.StatusCode(),
		Success:    resp.IsSuccess(),
	}, nil
}
