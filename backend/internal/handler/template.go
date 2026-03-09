package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lobsterpool/lobsterpool/internal/models"
)

type TemplateHandler struct {
	store *models.TemplateStore
}

func NewTemplateHandler(store *models.TemplateStore) *TemplateHandler {
	return &TemplateHandler{store: store}
}

func applyTemplateDefaults(tmpl *models.ClawTemplate) {
	if tmpl.Version == "" {
		tmpl.Version = "latest"
	}

	if tmpl.DefaultPort == 0 {
		tmpl.DefaultPort = 8080
	}
}

func (h *TemplateHandler) List(c *gin.Context) {
	templates, err := h.store.List()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list templates")
		return
	}
	if templates == nil {
		templates = []models.ClawTemplate{}
	}
	c.JSON(http.StatusOK, templates)
}

func (h *TemplateHandler) Get(c *gin.Context) {
	id := c.Param("id")
	tmpl, err := h.store.Get(id)
	switch {
	case err == sql.ErrNoRows:
		respondError(c, http.StatusNotFound, "template not found")
		return
	case err != nil:
		respondError(c, http.StatusInternalServerError, "failed to get template")
		return
	}
	c.JSON(http.StatusOK, tmpl)
}

func (h *TemplateHandler) Create(c *gin.Context) {
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

	if err := h.store.Create(&tmpl); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create template")
		return
	}

	created, _ := h.store.Get(tmpl.ID)
	c.JSON(http.StatusCreated, created)
}
