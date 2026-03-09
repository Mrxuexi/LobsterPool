package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lobsterpool/lobsterpool/internal/models"
	"github.com/lobsterpool/lobsterpool/internal/provider"
)

type InstanceHandler struct {
	instances *models.InstanceStore
	templates *models.TemplateStore
	users     *models.UserStore
	provider  provider.Provider
}

func NewInstanceHandler(instances *models.InstanceStore, templates *models.TemplateStore, users *models.UserStore, p provider.Provider) *InstanceHandler {
	return &InstanceHandler{
		instances: instances,
		templates: templates,
		users:     users,
		provider:  p,
	}
}

type CreateInstanceRequest struct {
	Name       string `json:"name" binding:"required"`
	TemplateID string `json:"template_id" binding:"required"`
	Cluster    string `json:"cluster"`
	APIKey     string `json:"api_key" binding:"required"`
	MMBotToken string `json:"mm_bot_token" binding:"required"`
}

func newInstanceID() string {
	return strings.Split(uuid.New().String(), "-")[0]
}

func resourceName(instanceID string) string {
	return fmt.Sprintf("claw-%s", instanceID)
}

func (h *InstanceHandler) newInstance(userID string, req CreateInstanceRequest, cluster provider.ClusterInfo) *models.Instance {
	instanceID := newInstanceID()
	name := resourceName(instanceID)

	return &models.Instance{
		ID:             instanceID,
		Name:           req.Name,
		TemplateID:     req.TemplateID,
		UserID:         userID,
		Cluster:        cluster.Name,
		Namespace:      cluster.Namespace,
		DeploymentName: name,
		ServiceName:    name,
		Status:         "pending",
	}
}

func (h *InstanceHandler) resolveCluster(requestedCluster string) (provider.ClusterInfo, bool) {
	clusters := h.provider.ListClusters()
	if len(clusters) == 0 {
		return provider.ClusterInfo{}, false
	}

	if requestedCluster == "" {
		for _, cluster := range clusters {
			if cluster.Default {
				return cluster, true
			}
		}
		return clusters[0], true
	}

	for _, cluster := range clusters {
		if cluster.Name == requestedCluster {
			return cluster, true
		}
	}
	return provider.ClusterInfo{}, false
}

func (h *InstanceHandler) refreshInstanceStatus(inst *models.Instance) {
	status, err := h.provider.GetInstanceStatus(inst)
	if err != nil {
		return
	}

	inst.Status = status.Status
	inst.Endpoint = status.Endpoint
}

func (h *InstanceHandler) Create(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok || userID == "" {
		respondError(c, http.StatusUnauthorized, authUnauthorizedError)
		return
	}

	var req CreateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request: name, template_id, api_key, and mm_bot_token are required")
		return
	}

	cluster, ok := h.resolveCluster(req.Cluster)
	if !ok {
		respondError(c, http.StatusBadRequest, "cluster not found")
		return
	}

	user, err := h.users.GetByID(userID)
	switch {
	case err == sql.ErrNoRows:
		respondError(c, http.StatusUnauthorized, authUserNotFoundError)
		return
	case err != nil:
		respondError(c, http.StatusInternalServerError, "failed to load current user")
		return
	}

	if user.MaxInstances > 0 {
		instanceCount, err := h.instances.CountByUser(userID)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "failed to count user instances")
			return
		}
		if instanceCount >= user.MaxInstances {
			respondError(c, http.StatusForbidden, "instance limit reached")
			return
		}
	}

	tmpl, err := h.templates.Get(req.TemplateID)
	switch {
	case err == sql.ErrNoRows:
		respondError(c, http.StatusBadRequest, "template not found")
		return
	case err != nil:
		respondError(c, http.StatusInternalServerError, "failed to get template")
		return
	}

	inst := h.newInstance(userID, req, cluster)

	if err := h.provider.CreateInstance(&provider.CreateInstanceInput{
		Instance:   inst,
		Template:   tmpl,
		APIKey:     req.APIKey,
		MMBotToken: req.MMBotToken,
	}); err != nil {
		respondError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create k8s resources: %v", err))
		return
	}

	h.refreshInstanceStatus(inst)

	if err := h.instances.Create(inst); err != nil {
		_ = h.provider.DeleteInstance(inst)
		respondError(c, http.StatusInternalServerError, "failed to save instance")
		return
	}

	c.JSON(http.StatusCreated, inst)
}

func (h *InstanceHandler) List(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok || userID == "" {
		respondError(c, http.StatusUnauthorized, authUnauthorizedError)
		return
	}

	instances, err := h.instances.ListByUser(userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list instances")
		return
	}
	if instances == nil {
		instances = []models.Instance{}
	}
	c.JSON(http.StatusOK, instances)
}

func (h *InstanceHandler) Get(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok || userID == "" {
		respondError(c, http.StatusUnauthorized, authUnauthorizedError)
		return
	}

	id := c.Param("id")
	inst, err := h.instances.GetByUser(id, userID)
	switch {
	case err == sql.ErrNoRows:
		respondError(c, http.StatusNotFound, "instance not found")
		return
	case err != nil:
		respondError(c, http.StatusInternalServerError, "failed to get instance")
		return
	}

	h.refreshInstanceStatus(inst)
	_ = h.instances.UpdateStatus(inst.ID, inst.Status, inst.Endpoint)

	c.JSON(http.StatusOK, inst)
}

func (h *InstanceHandler) Delete(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok || userID == "" {
		respondError(c, http.StatusUnauthorized, authUnauthorizedError)
		return
	}

	id := c.Param("id")
	inst, err := h.instances.GetByUser(id, userID)
	switch {
	case err == sql.ErrNoRows:
		respondError(c, http.StatusNotFound, "instance not found")
		return
	case err != nil:
		respondError(c, http.StatusInternalServerError, "failed to get instance")
		return
	}

	if err := h.provider.DeleteInstance(inst); err != nil {
		respondError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete k8s resources: %v", err))
		return
	}

	if err := h.instances.DeleteByUser(id, userID); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete instance record")
		return
	}

	respondMessage(c, http.StatusOK, "instance deleted")
}
