package models

import (
	"database/sql"
	"time"
)

const (
	UserRoleMember = "member"
	UserRoleAdmin  = "admin"
)

type User struct {
	ID                 string    `json:"id"`
	Username           string    `json:"username"`
	Role               string    `json:"role"`
	MaxInstances       int       `json:"max_instances"`
	MustChangePassword bool      `json:"must_change_password"`
	PasswordHash       string    `json:"-"`
	CreatedAt          time.Time `json:"created_at"`
}

type UserSummary struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	Role          string    `json:"role"`
	MaxInstances  int       `json:"max_instances"`
	InstanceCount int       `json:"instance_count"`
	CreatedAt     time.Time `json:"created_at"`
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(user *User) error {
	if user.Role == "" {
		user.Role = UserRoleMember
	}
	if user.Role == UserRoleMember && user.MaxInstances <= 0 {
		user.MaxInstances = 1
	}
	if user.Role == UserRoleAdmin && user.MaxInstances < 0 {
		user.MaxInstances = 0
	}

	_, err := s.db.Exec(
		`INSERT INTO users (id, username, password_hash, role, max_instances, must_change_password) VALUES (?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.PasswordHash, user.Role, user.MaxInstances, user.MustChangePassword,
	)
	return err
}

func (s *UserStore) GetByUsername(username string) (*User, error) {
	var user User
	err := s.db.QueryRow(
		`SELECT id, username, role, max_instances, must_change_password, password_hash, created_at FROM users WHERE username = ?`, username,
	).Scan(&user.ID, &user.Username, &user.Role, &user.MaxInstances, &user.MustChangePassword, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) GetByID(id string) (*User, error) {
	var user User
	err := s.db.QueryRow(
		`SELECT id, username, role, max_instances, must_change_password, password_hash, created_at FROM users WHERE id = ?`, id,
	).Scan(&user.ID, &user.Username, &user.Role, &user.MaxInstances, &user.MustChangePassword, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) UpdatePassword(id, passwordHash string, mustChangePassword bool) error {
	_, err := s.db.Exec(
		`UPDATE users SET password_hash = ?, must_change_password = ? WHERE id = ?`,
		passwordHash, mustChangePassword, id,
	)
	return err
}

func (s *UserStore) EnsureRoleByUsername(username, role string) error {
	_, err := s.db.Exec(`UPDATE users SET role = ? WHERE username = ?`, role, username)
	return err
}

func (s *UserStore) UpdateMaxInstances(id string, maxInstances int) error {
	_, err := s.db.Exec(`UPDATE users SET max_instances = ? WHERE id = ?`, maxInstances, id)
	return err
}

func (s *UserStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (s *UserStore) CountByRole(role string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = ?`, role).Scan(&count)
	return count, err
}

func (s *UserStore) ListSummaries(limit int) ([]UserSummary, error) {
	query := `
		SELECT
			u.id,
			u.username,
			u.role,
			u.max_instances,
			COUNT(i.id) AS instance_count,
			u.created_at
		FROM users u
		LEFT JOIN instances i ON i.user_id = u.id
		GROUP BY u.id, u.username, u.role, u.max_instances, u.created_at
		ORDER BY u.created_at DESC
	`

	var (
		rows *sql.Rows
		err  error
	)
	if limit > 0 {
		rows, err = s.db.Query(query+` LIMIT ?`, limit)
	} else {
		rows, err = s.db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserSummary
	for rows.Next() {
		var user UserSummary
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.MaxInstances, &user.InstanceCount, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}
