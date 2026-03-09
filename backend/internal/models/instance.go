package models

import (
	"database/sql"
	"time"
)

type Instance struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	TemplateID     string    `json:"template_id"`
	UserID         string    `json:"user_id,omitempty"`
	Cluster        string    `json:"cluster"`
	Namespace      string    `json:"namespace"`
	DeploymentName string    `json:"deployment_name"`
	ServiceName    string    `json:"service_name"`
	Status         string    `json:"status"`
	Endpoint       string    `json:"endpoint"`
	CreatedAt      time.Time `json:"created_at"`
}

type InstanceSummary struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	TemplateID     string    `json:"template_id"`
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	Cluster        string    `json:"cluster"`
	Namespace      string    `json:"namespace"`
	DeploymentName string    `json:"deployment_name"`
	ServiceName    string    `json:"service_name"`
	Status         string    `json:"status"`
	Endpoint       string    `json:"endpoint"`
	CreatedAt      time.Time `json:"created_at"`
}

type InstanceStore struct {
	db *sql.DB
}

func NewInstanceStore(db *sql.DB) *InstanceStore {
	return &InstanceStore{db: db}
}

func (s *InstanceStore) Create(inst *Instance) error {
	_, err := s.db.Exec(
		`INSERT INTO instances (id, name, template_id, user_id, cluster, namespace, deployment_name, service_name, status, endpoint)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		inst.ID, inst.Name, inst.TemplateID, inst.UserID, inst.Cluster, inst.Namespace,
		inst.DeploymentName, inst.ServiceName, inst.Status, inst.Endpoint,
	)
	return err
}

func (s *InstanceStore) ListByUser(userID string) ([]Instance, error) {
	rows, err := s.db.Query(
		`SELECT id, name, template_id, user_id, cluster, namespace, deployment_name, service_name, status, endpoint, created_at
		 FROM instances
		 WHERE user_id = ?
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []Instance
	for rows.Next() {
		var inst Instance
		if err := rows.Scan(&inst.ID, &inst.Name, &inst.TemplateID, &inst.UserID, &inst.Cluster, &inst.Namespace,
			&inst.DeploymentName, &inst.ServiceName, &inst.Status, &inst.Endpoint, &inst.CreatedAt); err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

func (s *InstanceStore) GetByUser(id, userID string) (*Instance, error) {
	var inst Instance
	err := s.db.QueryRow(
		`SELECT id, name, template_id, user_id, cluster, namespace, deployment_name, service_name, status, endpoint, created_at
		 FROM instances
		 WHERE id = ? AND user_id = ?`,
		id, userID,
	).Scan(&inst.ID, &inst.Name, &inst.TemplateID, &inst.UserID, &inst.Cluster, &inst.Namespace,
		&inst.DeploymentName, &inst.ServiceName, &inst.Status, &inst.Endpoint, &inst.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

func (s *InstanceStore) UpdateStatus(id, status, endpoint string) error {
	_, err := s.db.Exec("UPDATE instances SET status = ?, endpoint = ? WHERE id = ?", status, endpoint, id)
	return err
}

func (s *InstanceStore) AssignDefaultCluster(cluster string) error {
	_, err := s.db.Exec(
		`UPDATE instances SET cluster = ? WHERE cluster IS NULL OR TRIM(cluster) = ''`,
		cluster,
	)
	return err
}

func (s *InstanceStore) DeleteByUser(id, userID string) error {
	_, err := s.db.Exec("DELETE FROM instances WHERE id = ? AND user_id = ?", id, userID)
	return err
}

func (s *InstanceStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM instances").Scan(&count)
	return count, err
}

func (s *InstanceStore) CountRunning() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM instances WHERE status = ?", "running").Scan(&count)
	return count, err
}

func (s *InstanceStore) CountByUser(userID string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM instances WHERE user_id = ?", userID).Scan(&count)
	return count, err
}

func (s *InstanceStore) ListSummaries(limit int) ([]InstanceSummary, error) {
	query := `
		SELECT
			i.id,
			i.name,
			i.template_id,
			i.user_id,
			COALESCE(u.username, '') AS username,
			i.cluster,
			i.namespace,
			i.deployment_name,
			i.service_name,
			i.status,
			i.endpoint,
			i.created_at
		FROM instances i
		LEFT JOIN users u ON u.id = i.user_id
		ORDER BY i.created_at DESC
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

	var instances []InstanceSummary
	for rows.Next() {
		var inst InstanceSummary
		if err := rows.Scan(
			&inst.ID,
			&inst.Name,
			&inst.TemplateID,
			&inst.UserID,
			&inst.Username,
			&inst.Cluster,
			&inst.Namespace,
			&inst.DeploymentName,
			&inst.ServiceName,
			&inst.Status,
			&inst.Endpoint,
			&inst.CreatedAt,
		); err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}

	return instances, rows.Err()
}
