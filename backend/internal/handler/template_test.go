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

func setupTemplateHandler(t *testing.T) (*TemplateHandler, *models.TemplateStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := database.Open(filepath.Join(t.TempDir(), "handler-template.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store := models.NewTemplateStore(db)
	return NewTemplateHandler(store), store
}

func TestTemplateHandler_GetNotFound(t *testing.T) {
	h, _ := setupTemplateHandler(t)

	r := gin.New()
	r.GET("/templates/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/templates/not-exist", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestTemplateHandler_CreateAppliesDefaults(t *testing.T) {
	h, store := setupTemplateHandler(t)

	r := gin.New()
	r.POST("/templates", h.Create)

	body := map[string]any{
		"id":    "tmpl-defaults",
		"name":  "Template Defaults",
		"image": "repo/image",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	created, err := store.Get("tmpl-defaults")
	if err != nil {
		t.Fatalf("failed getting created template: %v", err)
	}
	if created.Version != "latest" {
		t.Fatalf("expected default version latest, got %q", created.Version)
	}
	if created.DefaultPort != 8080 {
		t.Fatalf("expected default port 8080, got %d", created.DefaultPort)
	}
}
