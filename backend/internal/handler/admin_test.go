package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lobsterpool/lobsterpool/internal/database"
	"github.com/lobsterpool/lobsterpool/internal/models"
)

func setupAdminHandler(t *testing.T) (*AdminHandler, *models.UserStore, *models.InstanceStore, *models.TemplateStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := database.Open(filepath.Join(t.TempDir(), "handler-admin.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userStore := models.NewUserStore(db)
	instStore := models.NewInstanceStore(db)
	templateStore := models.NewTemplateStore(db)

	return NewAdminHandler(userStore, templateStore, instStore), userStore, instStore, templateStore
}

func TestAdminHandler_Overview(t *testing.T) {
	h, userStore, instStore, _ := setupAdminHandler(t)

	if err := userStore.Create(&models.User{
		ID:           "user-1",
		Username:     "alice",
		Role:         models.UserRoleMember,
		PasswordHash: "hash",
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := instStore.Create(&models.Instance{
		ID:             "inst-1",
		Name:           "Workspace A",
		TemplateID:     "openclaw-mm",
		UserID:         "user-1",
		Namespace:      "ns",
		DeploymentName: "dep",
		ServiceName:    "svc",
		Status:         "running",
		Endpoint:       "http://endpoint",
	}); err != nil {
		t.Fatalf("create instance: %v", err)
	}

	r := gin.New()
	r.GET("/admin/overview", h.Overview)

	req := httptest.NewRequest(http.MethodGet, "/admin/overview", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var overview AdminOverviewResponse
	if err := json.Unmarshal(w.Body.Bytes(), &overview); err != nil {
		t.Fatalf("unmarshal overview: %v", err)
	}
	if overview.TotalUsers != 2 || overview.AdminUsers != 1 || overview.TotalInstances != 1 || overview.RunningInstances != 1 {
		t.Fatalf("unexpected overview: %+v", overview)
	}
}

func TestAdminHandler_CreateTemplate(t *testing.T) {
	h, _, _, templateStore := setupAdminHandler(t)

	r := gin.New()
	r.POST("/admin/templates", h.CreateTemplate)

	payload := map[string]any{
		"id":          "tmpl-admin",
		"name":        "Admin Template",
		"description": "managed",
		"image":       "repo/admin",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/admin/templates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}

	created, err := templateStore.Get("tmpl-admin")
	if err != nil {
		t.Fatalf("get template: %v", err)
	}
	if created.Version != "latest" || created.DefaultPort != 8080 {
		t.Fatalf("unexpected template defaults: %+v", created)
	}
}

func TestAdminHandler_UpdateUserMaxInstances(t *testing.T) {
	h, userStore, _, _ := setupAdminHandler(t)

	if err := userStore.Create(&models.User{
		ID:           "user-1",
		Username:     "alice",
		Role:         models.UserRoleMember,
		PasswordHash: "hash",
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	r := gin.New()
	r.PATCH("/admin/users/:id/max-instances", h.UpdateUserMaxInstances)

	body := []byte(`{"max_instances":3}`)
	req := httptest.NewRequest(http.MethodPatch, "/admin/users/user-1/max-instances", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	updatedUser, err := userStore.GetByID("user-1")
	if err != nil {
		t.Fatalf("get updated user: %v", err)
	}
	if updatedUser.MaxInstances != 3 {
		t.Fatalf("expected max_instances=3, got %+v", updatedUser)
	}
}
