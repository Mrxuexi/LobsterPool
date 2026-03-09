package database

import (
	"path/filepath"
	"testing"
)

func TestOpen_RunsMigrationsAndSeed(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "lobsterpool.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	for _, table := range []string{"users", "claw_templates", "instances"} {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("failed checking table %s: %v", table, err)
		}
		if count != 1 {
			t.Fatalf("expected table %s to exist", table)
		}
	}

	var templateCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM claw_templates").Scan(&templateCount); err != nil {
		t.Fatalf("failed counting templates: %v", err)
	}
	if templateCount == 0 {
		t.Fatalf("expected seed data to be inserted")
	}

	var adminCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount); err != nil {
		t.Fatalf("failed counting admins: %v", err)
	}
	if adminCount != 0 {
		t.Fatalf("expected no bootstrap admin by default, got %d", adminCount)
	}
}

func TestMigrate_IsIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "lobsterpool.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("second migrate failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM claw_templates WHERE id = 'openclaw-mm'").Scan(&count); err != nil {
		t.Fatalf("failed counting seeded template: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one seeded template after re-migrate, got %d", count)
	}
}

func TestEnsureBootstrapAdminUser_CreatesAdminWhenExplicitPasswordProvided(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "lobsterpool.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if err := EnsureBootstrapAdminUser(db, "bootstrap-admin", "bootstrap-secret"); err != nil {
		t.Fatalf("EnsureBootstrapAdminUser failed: %v", err)
	}

	var (
		username           string
		role               string
		maxInstances       int
		mustChangePassword bool
	)
	if err := db.QueryRow("SELECT username, role, max_instances, must_change_password FROM users WHERE role = 'admin'").Scan(
		&username,
		&role,
		&maxInstances,
		&mustChangePassword,
	); err != nil {
		t.Fatalf("failed loading bootstrapped admin: %v", err)
	}

	if username != "bootstrap-admin" || role != "admin" || maxInstances != 0 || !mustChangePassword {
		t.Fatalf("unexpected bootstrapped admin state: username=%s role=%s max=%d mustChange=%v", username, role, maxInstances, mustChangePassword)
	}
}

func TestEnsureBootstrapAdminUser_NoOpWithoutPassword(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "lobsterpool.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if err := EnsureBootstrapAdminUser(db, "admin", ""); err != nil {
		t.Fatalf("EnsureBootstrapAdminUser failed: %v", err)
	}

	var adminCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount); err != nil {
		t.Fatalf("failed counting admins: %v", err)
	}
	if adminCount != 0 {
		t.Fatalf("expected no admins without explicit bootstrap password, got %d", adminCount)
	}
}
