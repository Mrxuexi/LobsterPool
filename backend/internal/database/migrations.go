package database

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "admin"
)

func Migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'member',
			max_instances INTEGER NOT NULL DEFAULT 1,
			must_change_password INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS claw_templates (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			image TEXT NOT NULL,
			version TEXT NOT NULL DEFAULT 'latest',
			default_port INTEGER NOT NULL DEFAULT 8080,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS instances (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			template_id TEXT NOT NULL,
			user_id TEXT NOT NULL DEFAULT '',
			namespace TEXT NOT NULL,
			deployment_name TEXT NOT NULL,
			service_name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			endpoint TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (template_id) REFERENCES claw_templates(id)
		)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	if err := migrateUserRoles(db); err != nil {
		return fmt.Errorf("user role migration failed: %w", err)
	}

	if err := migrateMustChangePassword(db); err != nil {
		return fmt.Errorf("must-change-password migration failed: %w", err)
	}

	if err := migrateMaxInstances(db); err != nil {
		return fmt.Errorf("max-instances migration failed: %w", err)
	}

	if err := ensureDefaultAdminUser(db); err != nil {
		return fmt.Errorf("default admin setup failed: %w", err)
	}

	if err := seed(db); err != nil {
		return fmt.Errorf("seed failed: %w", err)
	}

	return nil
}

func migrateUserRoles(db *sql.DB) error {
	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'member'`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}

	if _, err := db.Exec(`UPDATE users SET role = 'member' WHERE role IS NULL OR TRIM(role) = ''`); err != nil {
		return err
	}

	return nil
}

func migrateMustChangePassword(db *sql.DB) error {
	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN must_change_password INTEGER NOT NULL DEFAULT 0`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}

	_, err := db.Exec(`UPDATE users SET must_change_password = 0 WHERE must_change_password IS NULL`)
	return err
}

func migrateMaxInstances(db *sql.DB) error {
	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN max_instances INTEGER NOT NULL DEFAULT 1`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}

	_, err := db.Exec(`UPDATE users SET max_instances = 1 WHERE max_instances IS NULL OR max_instances < 0`)
	return err
}

func ensureDefaultAdminUser(db *sql.DB) error {
	var existingID string
	err := db.QueryRow(`SELECT id FROM users WHERE username = ?`, defaultAdminUsername).Scan(&existingID)
	switch {
	case err == nil:
		_, updateErr := db.Exec(
			`UPDATE users SET role = 'admin', max_instances = 0 WHERE username = ?`,
			defaultAdminUsername,
		)
		return updateErr
	case err != sql.ErrNoRows:
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO users (id, username, password_hash, role, max_instances, must_change_password) VALUES (?, ?, ?, 'admin', 0, 1)`,
		"default-admin",
		defaultAdminUsername,
		string(hashedPassword),
	)
	return err
}

func seed(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM claw_templates").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	seeds := []struct {
		id, name, desc, image, version string
		port                           int
	}{
		{
			id:      "openclaw-mm",
			name:    "Mattermost Bot",
			desc:    "OpenClaw Mattermost Bot instance",
			image:   "registry.company.com/openclaw/mm-bot",
			version: "1.0",
			port:    8080,
		},
	}

	stmt, err := db.Prepare("INSERT INTO claw_templates (id, name, description, image, version, default_port) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range seeds {
		if _, err := stmt.Exec(s.id, s.name, s.desc, s.image, s.version, s.port); err != nil {
			return err
		}
	}
	return nil
}
