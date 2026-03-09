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

func setupAuthHandler(t *testing.T) *AuthHandler {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := database.Open(filepath.Join(t.TempDir(), "handler-auth.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return NewAuthHandler(models.NewUserStore(db), "test-secret")
}

func TestAuthHandler_RegisterAndLogin(t *testing.T) {
	h := setupAuthHandler(t)
	r := gin.New()
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)

	registerBody := map[string]string{"username": "alice", "password": "pwd123"}
	registerPayload, _ := json.Marshal(registerBody)
	registerReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(registerPayload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	r.ServeHTTP(registerW, registerReq)

	if registerW.Code != http.StatusCreated {
		t.Fatalf("expected register 201, got %d body=%s", registerW.Code, registerW.Body.String())
	}

	loginBody := map[string]string{"username": "alice", "password": "pwd123"}
	loginPayload, _ := json.Marshal(loginBody)
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d body=%s", loginW.Code, loginW.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(loginW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected non-empty token in login response")
	}
	user, ok := resp["user"].(map[string]any)
	if !ok {
		t.Fatalf("expected user object in login response")
	}
	if role, ok := user["role"].(string); !ok || role != models.UserRoleMember {
		t.Fatalf("expected registered user to be member, got %+v", user)
	}
	if maxInstances, ok := user["max_instances"].(float64); !ok || int(maxInstances) != 1 {
		t.Fatalf("expected registered user max_instances=1, got %+v", user)
	}
}

func TestAuthHandler_DefaultAdminLoginAndPasswordChange(t *testing.T) {
	h := setupAuthHandler(t)
	r := gin.New()
	r.POST("/auth/login", h.Login)
	r.POST("/auth/change-password", func(c *gin.Context) {
		c.Set("userID", "default-admin")
		h.ChangePassword(c)
	})

	loginPayload := []byte(`{"username":"admin","password":"admin"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("expected default admin login 200, got %d body=%s", loginW.Code, loginW.Body.String())
	}

	var loginResp map[string]any
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	user, ok := loginResp["user"].(map[string]any)
	if !ok {
		t.Fatalf("expected user object in login response")
	}
	if mustChange, ok := user["must_change_password"].(bool); !ok || !mustChange {
		t.Fatalf("expected default admin to require password change, got %+v", user)
	}
	if maxInstances, ok := user["max_instances"].(float64); !ok || int(maxInstances) != 0 {
		t.Fatalf("expected default admin max_instances=0, got %+v", user)
	}

	changePayload := []byte(`{"new_password":"new-admin-password"}`)
	changeReq := httptest.NewRequest(http.MethodPost, "/auth/change-password", bytes.NewReader(changePayload))
	changeReq.Header.Set("Content-Type", "application/json")
	changeW := httptest.NewRecorder()
	r.ServeHTTP(changeW, changeReq)

	if changeW.Code != http.StatusOK {
		t.Fatalf("expected password change 200, got %d body=%s", changeW.Code, changeW.Body.String())
	}

	newLoginPayload := []byte(`{"username":"admin","password":"new-admin-password"}`)
	newLoginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(newLoginPayload))
	newLoginReq.Header.Set("Content-Type", "application/json")
	newLoginW := httptest.NewRecorder()
	r.ServeHTTP(newLoginW, newLoginReq)

	if newLoginW.Code != http.StatusOK {
		t.Fatalf("expected login with updated password 200, got %d body=%s", newLoginW.Code, newLoginW.Body.String())
	}
}

func TestAuthHandler_LoginInvalidPassword(t *testing.T) {
	h := setupAuthHandler(t)
	r := gin.New()
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)

	registerPayload := []byte(`{"username":"bob","password":"abc"}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(registerPayload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	r.ServeHTTP(registerW, registerReq)
	if registerW.Code != http.StatusCreated {
		t.Fatalf("register failed: %d body=%s", registerW.Code, registerW.Body.String())
	}

	badLoginPayload := []byte(`{"username":"bob","password":"wrong"}`)
	badLoginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(badLoginPayload))
	badLoginReq.Header.Set("Content-Type", "application/json")
	badLoginW := httptest.NewRecorder()
	r.ServeHTTP(badLoginW, badLoginReq)

	if badLoginW.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", badLoginW.Code, badLoginW.Body.String())
	}
}
