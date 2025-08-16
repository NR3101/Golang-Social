package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/NR3101/social/internal/store"
	"github.com/redis/go-redis/v9"
)

// UserExpTime defines the expiration time for user cache entries
const UserExpTime = time.Minute

// UserStore implements the Users interface for Redis operations
type UserStore struct {
	rdb *redis.Client // Redis client for database operations
}

// Get retrieves a user by ID from the Redis cache
func (s *UserStore) Get(ctx context.Context, userID string) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// If the key does not exist, return nil without an error
			return nil, nil
		}
		return nil, err // Return any other error encountered
	}

	var user store.User
	if data != "" {
		if err := json.Unmarshal([]byte(data), &user); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

// Set stores a user in the Redis cache
func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	if err := s.rdb.SetEx(ctx, cacheKey, data, UserExpTime).Err(); err != nil {
		return err
	}

	return nil
}
