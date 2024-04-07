package config

type PagerDutyConfig struct {
	APIKey         string `json:"api_key"`
	IntegrationKey string `json:"integration_key"`
	IntegrationURL string `json:"integration_url"`
}
