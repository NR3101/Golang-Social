package store

import (
	"context"
	"database/sql"
	"errors"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
}

// Implements the Storage interface for roles
type RoleStore struct {
	db *sql.DB
}

// GetByName retrieves a role by its name from the database.
func (s *RoleStore) GetByName(ctx context.Context, name string) (*Role, error) {
	query := "SELECT id, name, description, level FROM roles WHERE name = $1"
	row := s.db.QueryRowContext(ctx, query, name)

	role := &Role{}
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.Level)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound // Role not found
		}
		return nil, err // Other error
	}

	return role, nil
}
