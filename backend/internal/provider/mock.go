package provider

import (
	"fmt"
	"log"
	"sync"

	"github.com/lobsterpool/lobsterpool/internal/models"
)

type MockProvider struct {
	mu        sync.RWMutex
	instances map[string]string // id -> status
}

const mockEndpointPort = 8080

func NewMockProvider() *MockProvider {
	return &MockProvider{
		instances: make(map[string]string),
	}
}

func (m *MockProvider) CreateInstance(input *CreateInstanceInput) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[MockProvider] Creating instance %s (template: %s, image: %s:%s)",
		input.Instance.ID, input.Template.ID, input.Template.Image, input.Template.Version)
	log.Printf("[MockProvider] Secret would contain: API_KEY=***, MM_BOT_TOKEN=***")

	m.instances[input.Instance.ID] = "running"
	return nil
}

func (m *MockProvider) DeleteInstance(instance *models.Instance) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[MockProvider] Deleting instance %s (deployment: %s, service: %s)",
		instance.ID, instance.DeploymentName, instance.ServiceName)

	delete(m.instances, instance.ID)
	return nil
}

func (m *MockProvider) GetInstanceStatus(instance *models.Instance) (*InstanceStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, ok := m.instances[instance.ID]
	if !ok {
		return &InstanceStatus{Status: "not_found"}, nil
	}

	endpoint := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", instance.ServiceName, instance.Namespace, mockEndpointPort)

	return &InstanceStatus{
		Status:   status,
		Endpoint: endpoint,
	}, nil
}
