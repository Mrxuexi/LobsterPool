package provider

import "github.com/lobsterpool/lobsterpool/internal/models"

type CreateInstanceInput struct {
	Instance   *models.Instance
	Template   *models.ClawTemplate
	APIKey     string
	MMBotToken string
}

type InstanceStatus struct {
	Status   string `json:"status"`
	Endpoint string `json:"endpoint"`
}

type Provider interface {
	CreateInstance(input *CreateInstanceInput) error
	DeleteInstance(instance *models.Instance) error
	GetInstanceStatus(instance *models.Instance) (*InstanceStatus, error)
}
