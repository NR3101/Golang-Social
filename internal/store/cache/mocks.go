package cache

import (
	"context"

	"github.com/NR3101/social/internal/store"
)

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

// MockUserStore is a mock implementation of the UserStore interface for testing purposes.
type MockUserStore struct {
}

func (m *MockUserStore) Get(ctx context.Context, userID string) (*store.User, error) {
	return nil, nil
}

func (m *MockUserStore) Set(ctx context.Context, user *store.User) error {
	return nil
}
