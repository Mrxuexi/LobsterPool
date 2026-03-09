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

	var (
		adminRole          string
		mustChangePassword bool
	)
	if err := db.QueryRow("SELECT role, must_change_password FROM users WHERE username = 'admin'").Scan(&adminRole, &mustChangePassword); err != nil {
		t.Fatalf("failed loading default admin: %v", err)
	}
	if adminRole != "admin" {
		t.Fatalf("expected default admin role, got %q", adminRole)
	}
	if !mustChangePassword {
		t.Fatalf("expected default admin to require password change")
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
