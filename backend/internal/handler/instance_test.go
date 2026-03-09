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
	"github.com/lobsterpool/lobsterpool/internal/provider"
)

type fakeProvider struct {
	createErr   error
	deleteErr   error
	status      *provider.InstanceStatus
	createCalls int
	deleteCalls int
	lastInput   *provider.CreateInstanceInput
}

func (f *fakeProvider) CreateInstance(input *provider.CreateInstanceInput) error {
	f.createCalls++
	f.lastInput = input
	return f.createErr
}

func (f *fakeProvider) DeleteInstance(instance *models.Instance) error {
	f.deleteCalls++
	return f.deleteErr
}

func (f *fakeProvider) GetInstanceStatus(instance *models.Instance) (*provider.InstanceStatus, error) {
	if f.status != nil {
		return f.status, nil
	}
	return &provider.InstanceStatus{Status: "pending", Endpoint: ""}, nil
}

func setupInstanceHandler(t *testing.T, withUser bool, p provider.Provider) (*gin.Engine, *models.InstanceStore, *models.UserStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := database.Open(filepath.Join(t.TempDir(), "handler-instance.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	instStore := models.NewInstanceStore(db)
	tmplStore := models.NewTemplateStore(db)
	userStore := models.NewUserStore(db)
	h := NewInstanceHandler(instStore, tmplStore, userStore, p, "test-ns")

	if withUser {
		if err := userStore.Create(&models.User{
			ID:           "user-1",
			Username:     "alice",
			Role:         models.UserRoleMember,
			PasswordHash: "hash",
		}); err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	r := gin.New()
	if withUser {
		r.Use(func(c *gin.Context) {
			c.Set("userID", "user-1")
			c.Next()
		})
	}
	r.POST("/instances", h.Create)
	r.DELETE("/instances/:id", h.Delete)

	return r, instStore, userStore
}

func TestInstanceHandler_CreateUnauthorized(t *testing.T) {
	r, _, _ := setupInstanceHandler(t, false, &fakeProvider{})

	body := []byte(`{"name":"n1","template_id":"openclaw-mm","api_key":"k","mm_bot_token":"m"}`)
	req := httptest.NewRequest(http.MethodPost, "/instances", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestInstanceHandler_CreateSuccess(t *testing.T) {
	fp := &fakeProvider{status: &provider.InstanceStatus{Status: "running", Endpoint: "http://svc"}}
	r, instStore, _ := setupInstanceHandler(t, true, fp)

	payload := map[string]string{
		"name":         "My Claw",
		"template_id":  "openclaw-mm",
		"api_key":      "api-secret",
		"mm_bot_token": "mm-secret",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/instances", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	if fp.createCalls != 1 {
		t.Fatalf("expected provider create to be called once")
	}
	if fp.lastInput == nil || fp.lastInput.APIKey != "api-secret" || fp.lastInput.MMBotToken != "mm-secret" {
		t.Fatalf("provider received unexpected secrets: %+v", fp.lastInput)
	}

	instances, err := instStore.ListByUser("user-1")
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected one stored instance, got %d", len(instances))
	}
	if instances[0].Status != "running" || instances[0].Endpoint != "http://svc" {
		t.Fatalf("unexpected stored status: %+v", instances[0])
	}
}

func TestInstanceHandler_CreateLimitReached(t *testing.T) {
	fp := &fakeProvider{}
	r, instStore, _ := setupInstanceHandler(t, true, fp)

	if err := instStore.Create(&models.Instance{
		ID:             "inst-existing",
		Name:           "Existing",
		TemplateID:     "openclaw-mm",
		UserID:         "user-1",
		Namespace:      "test-ns",
		DeploymentName: "claw-inst-existing",
		ServiceName:    "claw-inst-existing",
		Status:         "running",
	}); err != nil {
		t.Fatalf("prepare instance: %v", err)
	}

	body := []byte(`{"name":"n2","template_id":"openclaw-mm","api_key":"k","mm_bot_token":"m"}`)
	req := httptest.NewRequest(http.MethodPost, "/instances", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", w.Code, w.Body.String())
	}
	if fp.createCalls != 0 {
		t.Fatalf("expected provider create not to be called")
	}
}

func TestInstanceHandler_DeleteSuccess(t *testing.T) {
	fp := &fakeProvider{}
	r, instStore, _ := setupInstanceHandler(t, true, fp)

	inst := &models.Instance{
		ID:             "inst-delete",
		Name:           "to-delete",
		TemplateID:     "openclaw-mm",
		UserID:         "user-1",
		Namespace:      "test-ns",
		DeploymentName: "claw-inst-delete",
		ServiceName:    "claw-inst-delete",
		Status:         "running",
	}
	if err := instStore.Create(inst); err != nil {
		t.Fatalf("prepare instance: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/instances/inst-delete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if fp.deleteCalls != 1 {
		t.Fatalf("expected provider delete to be called once")
	}
	if _, err := instStore.GetByUser("inst-delete", "user-1"); err == nil {
		t.Fatalf("expected instance to be deleted from db")
	}
}
