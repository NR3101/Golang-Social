package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyFollowing  = errors.New("already following this user")
	QueryTimeoutDuration = 5 * time.Second // Default timeout for database queries
)

// Storage defines the interface for the storage layer, which includes methods for managing various entities.
type Storage struct {
	// Posts provides methods for managing blog posts.
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, string) (*Post, error)
		Delete(context.Context, string) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostsForFeed, error) // Get posts for a specific user
	}

	// Users provides methods for managing users.
	Users interface {
		GetByID(context.Context, string) (*User, error)
		Create(context.Context, *sql.Tx, *User) error                        // Create a user
		Delete(context.Context, int64) error                                 // Delete a user
		CreateAndInvite(context.Context, *User, string, time.Duration) error // Create a user and send an invitation email
		Activate(context.Context, string) error                              // Activate a user account with a token
	}

	// Comments provides methods for managing comments.
	Comments interface {
		Create(context.Context, *Comment) error
		GetByPostID(context.Context, int64) ([]*Comment, error)
	}

	// Followers provides methods for managing user relationships.
	Followers interface {
		Follow(ctx context.Context, toFollowID int64, userID int64) error     // Follow another user
		Unfollow(ctx context.Context, toUnfollowID int64, userID int64) error // Unfollow another user
	}
}

// NewStorage creates a new Storage instance with the provided database connection.
func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		Users:     &UserStore{db},
		Comments:  &CommentStore{db},
		Followers: &FollowerStore{db},
	}
}

// withTx executes a function within a database transaction.
func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit()
}
