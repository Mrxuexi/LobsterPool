package models

import (
	"database/sql"
	"time"
)

type ClawTemplate struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Version     string    `json:"version"`
	DefaultPort int       `json:"default_port"`
	CreatedAt   time.Time `json:"created_at"`
}

type TemplateStore struct {
	db *sql.DB
}

func NewTemplateStore(db *sql.DB) *TemplateStore {
	return &TemplateStore{db: db}
}

func (s *TemplateStore) List() ([]ClawTemplate, error) {
	rows, err := s.db.Query("SELECT id, name, description, image, version, default_port, created_at FROM claw_templates ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []ClawTemplate
	for rows.Next() {
		var t ClawTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Image, &t.Version, &t.DefaultPort, &t.CreatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

func (s *TemplateStore) Get(id string) (*ClawTemplate, error) {
	var t ClawTemplate
	err := s.db.QueryRow("SELECT id, name, description, image, version, default_port, created_at FROM claw_templates WHERE id = ?", id).
		Scan(&t.ID, &t.Name, &t.Description, &t.Image, &t.Version, &t.DefaultPort, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TemplateStore) Create(t *ClawTemplate) error {
	_, err := s.db.Exec(
		"INSERT INTO claw_templates (id, name, description, image, version, default_port) VALUES (?, ?, ?, ?, ?, ?)",
		t.ID, t.Name, t.Description, t.Image, t.Version, t.DefaultPort,
	)
	return err
}

func (s *TemplateStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM claw_templates").Scan(&count)
	return count, err
}
