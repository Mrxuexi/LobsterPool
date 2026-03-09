package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lobsterpool/lobsterpool/internal/models"
)

type AdminHandler struct {
	users     *models.UserStore
	templates *models.TemplateStore
	instances *models.InstanceStore
}

type AdminOverviewResponse struct {
	TotalUsers       int                      `json:"total_users"`
	AdminUsers       int                      `json:"admin_users"`
	TotalInstances   int                      `json:"total_instances"`
	RunningInstances int                      `json:"running_instances"`
	TotalTemplates   int                      `json:"total_templates"`
	RecentUsers      []models.UserSummary     `json:"recent_users"`
	RecentInstances  []models.InstanceSummary `json:"recent_instances"`
}

type UpdateUserMaxInstancesRequest struct {
	MaxInstances int `json:"max_instances"`
}

func NewAdminHandler(
	userStore *models.UserStore,
	templateStore *models.TemplateStore,
	instanceStore *models.InstanceStore,
) *AdminHandler {
	return &AdminHandler{
		users:     userStore,
		templates: templateStore,
		instances: instanceStore,
	}
}

func (h *AdminHandler) Overview(c *gin.Context) {
	totalUsers, err := h.users.Count()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to count users")
		return
	}

	adminUsers, err := h.users.CountByRole(models.UserRoleAdmin)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to count admins")
		return
	}

	totalInstances, err := h.instances.Count()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to count instances")
		return
	}

	runningInstances, err := h.instances.CountRunning()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to count running instances")
		return
	}

	totalTemplates, err := h.templates.Count()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to count templates")
		return
	}

	recentUsers, err := h.users.ListSummaries(6)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list recent users")
		return
	}

	recentInstances, err := h.instances.ListSummaries(8)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list recent instances")
		return
	}

	c.JSON(http.StatusOK, AdminOverviewResponse{
		TotalUsers:       totalUsers,
		AdminUsers:       adminUsers,
		TotalInstances:   totalInstances,
		RunningInstances: runningInstances,
		TotalTemplates:   totalTemplates,
		RecentUsers:      recentUsers,
		RecentInstances:  recentInstances,
	})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.users.ListSummaries(0)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list users")
		return
	}
	if users == nil {
		users = []models.UserSummary{}
	}

	c.JSON(http.StatusOK, users)
}

func (h *AdminHandler) ListInstances(c *gin.Context) {
	instances, err := h.instances.ListSummaries(0)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list instances")
		return
	}
	if instances == nil {
		instances = []models.InstanceSummary{}
	}

	c.JSON(http.StatusOK, instances)
}

func (h *AdminHandler) UpdateUserMaxInstances(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		respondError(c, http.StatusBadRequest, "user id is required")
		return
	}

	var req UpdateUserMaxInstancesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.MaxInstances < 0 {
		respondError(c, http.StatusBadRequest, "max_instances must be greater than or equal to 0")
		return
	}

	if _, err := h.users.GetByID(userID); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondError(c, http.StatusNotFound, "user not found")
		default:
			respondError(c, http.StatusInternalServerError, "failed to load user")
		}
		return
	}

	if err := h.users.UpdateMaxInstances(userID, req.MaxInstances); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update user instance limit")
		return
	}

	updatedUser, err := h.users.GetByID(userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load updated user")
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

func (h *AdminHandler) CreateTemplate(c *gin.Context) {
	var tmpl models.ClawTemplate
	if err := c.ShouldBindJSON(&tmpl); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if tmpl.ID == "" || tmpl.Name == "" || tmpl.Image == "" {
		respondError(c, http.StatusBadRequest, "id, name, and image are required")
		return
	}

	applyTemplateDefaults(&tmpl)

	if err := h.templates.Create(&tmpl); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create template")
		return
	}

	created, err := h.templates.Get(tmpl.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load created template")
		return
	}

	c.JSON(http.StatusCreated, created)
}
