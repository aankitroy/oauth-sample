// internal/rbac/rbac.go
package rbac

import (
	"database/sql"
	"fmt"
)

type RBACStore struct {
	db *sql.DB
}

func NewRBACStore(db *sql.DB) *RBACStore {
	return &RBACStore{db: db}
}

func (s *RBACStore) GetUserRole(email string) (string, error) {
	var role string
	err := s.db.QueryRow(`SELECT role FROM users WHERE email=$1`, email).Scan(&role)
	if err != nil {
		return "", fmt.Errorf("failed to get user role: %w", err)
	}
	return role, nil
}
