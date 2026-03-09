package provider

import "github.com/lobsterpool/lobsterpool/internal/models"

type ClusterConfig struct {
	Name                  string
	DisplayName           string
	Namespace             string
	Kubeconfig            string
	Context               string
	APIServer             string
	Token                 string
	CAFile                string
	InsecureSkipTLSVerify bool
}

type ClusterInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Namespace   string `json:"namespace"`
	Default     bool   `json:"default"`
}

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
	ListClusters() []ClusterInfo
}
