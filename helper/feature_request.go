package helper

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"oryoo.com/models"
)

func InsertFeatureRequest(m *models.FeatureRequest) error {
	now := time.Now()
	m.CreatedAt = &now
	m.UpdatedAt = &now

	query := `
		INSERT INTO feature_requests (
			id,
			title,
			description,
			category,
			created_by,
			status,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := DB.ExecContext(
		context.Background(),
		query,
		m.ID,
		m.Title,
		m.Description,
		m.Category,
		m.CreatedBy,
		m.Status,
		m.CreatedAt,
		m.UpdatedAt,
	)
	return err
}

func scanFeatureRequest(scanner interface{ Scan(dest ...any) error }, m *models.FeatureRequest) error {
	var category sql.NullString
	err := scanner.Scan(
		&m.ID,
		&m.Title,
		&m.Description,
		&category,
		&m.CreatedBy,
		&m.Status,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if category.Valid {
		s := category.String
		m.Category = &s
	} else {
		m.Category = nil
	}
	return nil
}

const featureRequestSelectColumns = `
			id,
			title,
			description,
			category,
			created_by,
			status,
			created_at,
			updated_at`

func GetFeatureRequestByID(id string) (*models.FeatureRequest, error) {
	query := `SELECT ` + featureRequestSelectColumns + ` FROM feature_requests WHERE id = $1`
	var m models.FeatureRequest
	err := scanFeatureRequest(DB.QueryRowContext(context.Background(), query, id), &m)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("feature request not found")
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type FeatureRequestListFilter struct {
	CreatedBy string
	Category  string
	Status    string
}

func ListFeatureRequests(filter FeatureRequestListFilter) ([]models.FeatureRequest, error) {
	query := `SELECT ` + featureRequestSelectColumns + `
		FROM feature_requests
		WHERE created_by = $1`
	args := []interface{}{filter.CreatedBy}
	n := 2

	if c := strings.TrimSpace(filter.Category); c != "" {
		query += fmt.Sprintf(" AND LOWER(category) = LOWER($%d)", n)
		args = append(args, c)
		n++
	}
	if s := strings.TrimSpace(filter.Status); s != "" {
		query += fmt.Sprintf(" AND LOWER(status) = LOWER($%d)", n)
		args = append(args, s)
	}

	query += " ORDER BY created_at DESC"

	rows, err := DB.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.FeatureRequest, 0)
	for rows.Next() {
		var m models.FeatureRequest
		if err := scanFeatureRequest(rows, &m); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func UpdateFeatureRequest(m *models.FeatureRequest) error {
	now := time.Now()
	m.UpdatedAt = &now
	query := `
		UPDATE feature_requests SET
			title = $1,
			description = $2,
			category = $3,
			updated_at = $4
		WHERE id = $5 AND created_by = $6
	`
	res, err := DB.ExecContext(
		context.Background(),
		query,
		m.Title,
		m.Description,
		m.Category,
		m.UpdatedAt,
		m.ID,
		m.CreatedBy,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("feature request not found")
	}
	return nil
}

func DeleteFeatureRequest(id, createdBy string) error {
	res, err := DB.ExecContext(
		context.Background(),
		`DELETE FROM feature_requests WHERE id = $1 AND created_by = $2`,
		id,
		createdBy,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("feature request not found")
	}
	return nil
}
