package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

// FollowerStore implements the Storage interface for managing user follow relationships.
type FollowerStore struct {
	db *sql.DB
}

// Follower represents a follow relationship between two users.
type Follower struct {
	UserID         int64  `json:"user_id"`     // ID of the user who is following
	FollowedUserID int64  `json:"follower_id"` // ID of the user being followed
	CreatedAt      string `json:"created_at"`  // Timestamp when the follow relationship was created
}

// Follow allows a user to follow another user.
func (f *FollowerStore) Follow(ctx context.Context, toFollowID int64, userID int64) error {
	query := `INSERT INTO followers (user_id, follower_id) VALUES ($1, $2)`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := f.db.ExecContext(ctx, query, userID, toFollowID)
	if err != nil {
		// Check for duplicate key error (PostgreSQL error code 23505)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrAlreadyFollowing
		}
		return err
	}
	return nil
}

// Unfollow allows a user to unfollow another user.
func (f *FollowerStore) Unfollow(ctx context.Context, toUnfollowID int64, userID int64) error {
	query := `DELETE FROM followers WHERE user_id = $1 AND follower_id = $2`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := f.db.ExecContext(ctx, query, userID, toUnfollowID)
	return err
}
