package store

import (
	"context"
	"database/sql"
	"time"
)

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

// MockUserStore is a mock implementation of the UserStore interface for testing purposes.
type MockUserStore struct {
}

func (m *MockUserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	return nil
}

func (m *MockUserStore) GetByID(ctx context.Context, id string) (*User, error) {
	return &User{}, nil
}

func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	return &User{}, nil
}

func (m *MockUserStore) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *User, email string, duration time.Duration) error {
	return nil
}

func (m *MockUserStore) Activate(ctx context.Context, token string) error {
	return nil
}
