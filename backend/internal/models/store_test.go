package models

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/lobsterpool/lobsterpool/internal/database"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "models.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestTemplateStore_CreateAndGet(t *testing.T) {
	db := openTestDB(t)
	store := NewTemplateStore(db)

	tmpl := &ClawTemplate{
		ID:          "tmpl-1",
		Name:        "Template 1",
		Description: "desc",
		Image:       "repo/tmpl",
		Version:     "1.2.3",
		DefaultPort: 9090,
	}
	if err := store.Create(tmpl); err != nil {
		t.Fatalf("create template: %v", err)
	}

	got, err := store.Get("tmpl-1")
	if err != nil {
		t.Fatalf("get template: %v", err)
	}
	if got.Name != tmpl.Name || got.Image != tmpl.Image || got.DefaultPort != tmpl.DefaultPort {
		t.Fatalf("unexpected template: %+v", got)
	}
}

func TestInstanceStore_CRUDByUser(t *testing.T) {
	db := openTestDB(t)
	instStore := NewInstanceStore(db)

	inst := &Instance{
		ID:             "inst-1",
		Name:           "Instance 1",
		TemplateID:     "openclaw-mm",
		UserID:         "user-a",
		Namespace:      "ns",
		DeploymentName: "dep",
		ServiceName:    "svc",
		Status:         "pending",
		Endpoint:       "",
	}
	if err := instStore.Create(inst); err != nil {
		t.Fatalf("create instance: %v", err)
	}

	list, err := instStore.ListByUser("user-a")
	if err != nil {
		t.Fatalf("list by user: %v", err)
	}
	if len(list) != 1 || list[0].ID != "inst-1" {
		t.Fatalf("unexpected list result: %+v", list)
	}

	if err := instStore.UpdateStatus("inst-1", "running", "http://endpoint"); err != nil {
		t.Fatalf("update status: %v", err)
	}

	got, err := instStore.GetByUser("inst-1", "user-a")
	if err != nil {
		t.Fatalf("get by user: %v", err)
	}
	if got.Status != "running" || got.Endpoint != "http://endpoint" {
		t.Fatalf("unexpected status/endpoint: %+v", got)
	}

	if err := instStore.DeleteByUser("inst-1", "user-a"); err != nil {
		t.Fatalf("delete by user: %v", err)
	}

	if _, err := instStore.GetByUser("inst-1", "user-a"); err == nil {
		t.Fatalf("expected not found after delete")
	}
}

func TestInstanceStore_ListSummaries(t *testing.T) {
	db := openTestDB(t)
	userStore := NewUserStore(db)
	instStore := NewInstanceStore(db)

	user := &User{ID: "user-a", Username: "alice", Role: UserRoleAdmin, PasswordHash: "hash"}
	if err := userStore.Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	inst := &Instance{
		ID:             "inst-1",
		Name:           "Instance 1",
		TemplateID:     "openclaw-mm",
		UserID:         "user-a",
		Namespace:      "ns",
		DeploymentName: "dep",
		ServiceName:    "svc",
		Status:         "running",
		Endpoint:       "http://endpoint",
	}
	if err := instStore.Create(inst); err != nil {
		t.Fatalf("create instance: %v", err)
	}

	summaries, err := instStore.ListSummaries(0)
	if err != nil {
		t.Fatalf("list summaries: %v", err)
	}
	if len(summaries) < 1 {
		t.Fatalf("expected at least one summary, got %d", len(summaries))
	}

	foundAlice := false
	for _, summary := range summaries {
		if summary.Username == "alice" {
			foundAlice = true
			if summary.Status != "running" {
				t.Fatalf("unexpected summary: %+v", summary)
			}
		}
	}
	if !foundAlice {
		t.Fatalf("did not find alice in summaries: %+v", summaries)
	}
}

func TestUserStore_CreateAndGet(t *testing.T) {
	db := openTestDB(t)
	userStore := NewUserStore(db)

	user := &User{ID: "user-1", Username: "alice", Role: UserRoleAdmin, PasswordHash: "hash"}
	if err := userStore.Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	gotByUsername, err := userStore.GetByUsername("alice")
	if err != nil {
		t.Fatalf("get by username: %v", err)
	}
	if gotByUsername.ID != "user-1" {
		t.Fatalf("unexpected user: %+v", gotByUsername)
	}

	gotByID, err := userStore.GetByID("user-1")
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if gotByID.Username != "alice" {
		t.Fatalf("unexpected user: %+v", gotByID)
	}
	if gotByID.Role != UserRoleAdmin {
		t.Fatalf("expected admin role, got %+v", gotByID)
	}
	if gotByID.MustChangePassword {
		t.Fatalf("expected must_change_password to default to false")
	}
	if gotByID.MaxInstances != 0 {
		t.Fatalf("expected admin max_instances to default to 0, got %+v", gotByID)
	}
}

func TestUserStore_ListSummaries(t *testing.T) {
	db := openTestDB(t)
	userStore := NewUserStore(db)
	instStore := NewInstanceStore(db)

	users := []*User{
		{ID: "user-1", Username: "alice", Role: UserRoleAdmin, PasswordHash: "hash"},
		{ID: "user-2", Username: "bob", Role: UserRoleMember, PasswordHash: "hash"},
	}
	for _, user := range users {
		if err := userStore.Create(user); err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	inst := &Instance{
		ID:             "inst-1",
		Name:           "Instance 1",
		TemplateID:     "openclaw-mm",
		UserID:         "user-1",
		Namespace:      "ns",
		DeploymentName: "dep",
		ServiceName:    "svc",
		Status:         "pending",
	}
	if err := instStore.Create(inst); err != nil {
		t.Fatalf("create instance: %v", err)
	}

	summaries, err := userStore.ListSummaries(0)
	if err != nil {
		t.Fatalf("list summaries: %v", err)
	}
	if len(summaries) < 2 {
		t.Fatalf("expected at least two summaries, got %d", len(summaries))
	}

	foundAlice := false
	for _, summary := range summaries {
		if summary.Username == "alice" {
			foundAlice = true
			if summary.Role != UserRoleAdmin || summary.InstanceCount != 1 || summary.MaxInstances != 0 {
				t.Fatalf("unexpected alice summary: %+v", summary)
			}
		}
	}
	if !foundAlice {
		t.Fatalf("did not find alice in summaries: %+v", summaries)
	}
}

func TestUserStore_UpdateMaxInstances(t *testing.T) {
	db := openTestDB(t)
	userStore := NewUserStore(db)

	user := &User{ID: "user-1", Username: "alice", Role: UserRoleMember, PasswordHash: "hash"}
	if err := userStore.Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := userStore.UpdateMaxInstances("user-1", 3); err != nil {
		t.Fatalf("update max instances: %v", err)
	}

	updatedUser, err := userStore.GetByID("user-1")
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if updatedUser.MaxInstances != 3 {
		t.Fatalf("expected max_instances=3, got %+v", updatedUser)
	}
}
