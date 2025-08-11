package store

import (
	"context"
	"database/sql"
)

// Implementing the Storage interface for comments
type CommentStore struct {
	db *sql.DB
}

// CommentUser represents a user who created a comment.
type CommentUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// Comment represents a comment on a post.
type Comment struct {
	ID        int64       `json:"id"`
	PostID    int64       `json:"post_id"` // ID of the post this comment belongs to
	UserID    int64       `json:"user_id"` // ID of the user who created the comment
	Content   string      `json:"content"` // Content of the comment
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
	User      CommentUser `json:"user"` // User who created the comment
}

// Create inserts a new comment into the database.
func (c *CommentStore) Create(ctx context.Context, comment *Comment) error {
	query := `INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3) 
			  RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := c.db.QueryRowContext(ctx, query, comment.PostID, comment.UserID, comment.Content).
		Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// GetByPostID retrieves all comments for a specific post by its ID, along with the user information for each comment.
func (c *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]*Comment, error) {
	query := `SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, c.updated_at, u.id, u.username FROM 
			  comments c JOIN users u ON u.id = c.user_id 
              WHERE c.post_id = $1
              ORDER BY c.created_at DESC`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		comment := &Comment{}
		comment.User = CommentUser{} // Initialize User to avoid nil pointer dereference
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content,
			&comment.CreatedAt, &comment.UpdatedAt, &comment.User.ID, &comment.User.Username); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}
