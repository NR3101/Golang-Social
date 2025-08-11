package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

// Post represents a blog post in the system.
type Post struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	UserID    int64      `json:"user_id"` // ID of the user who created the post
	Tags      []string   `json:"tags"`    // Tags associated with the post
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
	Version   int64      `json:"version"`  // Version of the post for optimistic concurrency control
	Comments  []*Comment `json:"comments"` // Comments associated with the post
	User      *User      `json:"user"`     // User who created the post
}

// PostsForFeed represents a post with additional information for the user feed.
type PostsForFeed struct {
	Post
	CommentsCount int64 `json:"comments_count"` // Number of comments on the post
}

// Implementing the Storage interface for posts
type PostStore struct {
	db *sql.DB
}

// Create inserts a new post into the database.
func (p *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (title, content, user_id, tags) VALUES ($1, $2, $3, $4) 
			RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := p.db.QueryRowContext(ctx, query, post.Title, post.Content, post.UserID, pq.Array(post.Tags)).
		Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a post by its ID from the database with its associated comments.
func (p *PostStore) GetByID(ctx context.Context, postID string) (*Post, error) {
	query := `SELECT id, title, content, user_id, tags, created_at, updated_at, version FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	post := &Post{}
	err := p.db.QueryRowContext(ctx, query, postID).Scan(&post.ID, &post.Title, &post.Content,
		&post.UserID, pq.Array(&post.Tags), &post.CreatedAt, &post.UpdatedAt, &post.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return post, nil
}

// Delete removes a post by its ID from the database.
func (p *PostStore) Delete(ctx context.Context, postID string) error {
	query := `DELETE FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := p.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Update modifies an existing post in the database.
func (p *PostStore) Update(ctx context.Context, post *Post) error {
	query := `UPDATE posts SET title = $1, content = $2, version=version+1, updated_at = NOW() 
			WHERE id = $3 AND version=$4 RETURNING updated_at, version`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := p.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID, post.Version).
		Scan(&post.UpdatedAt, &post.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	return nil
}

// GetUserFeed retrieves posts for a specific user, including comments count.
func (p *PostStore) GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]PostsForFeed, error) {
	// Handle sort parameter safely
	sortDir := "DESC"
	if fq.Sort == "asc" {
		sortDir = "ASC"
	}

	// Handle empty tags - if no tags provided, don't filter by tags
	var tagsCondition string
	var queryArgs []interface{}

	if len(fq.Tags) > 0 {
		tagsCondition = "AND p.tags && $5"
		queryArgs = []interface{}{userID, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags)}
	} else {
		tagsCondition = ""
		queryArgs = []interface{}{userID, fq.Limit, fq.Offset, fq.Search}
	}

	query := `
   SELECT
    p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at, p.version, p.tags,
    u.username,
    COUNT(c.id) AS comments_count
   FROM posts p
   LEFT JOIN comments c ON c.post_id = p.id
   LEFT JOIN users u ON p.user_id = u.id
   LEFT JOIN followers f ON f.follower_id = p.user_id
   WHERE
    (f.user_id = $1 OR p.user_id = $1) AND
    ($4 = '' OR p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
    ` + tagsCondition + `
  GROUP BY p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at, p.version, p.tags, u.username
  ORDER BY p.created_at ` + sortDir + `
  LIMIT $2 OFFSET $3
 `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []PostsForFeed
	for rows.Next() {
		var post PostsForFeed
		post.User = &User{} // Initialize User pointer

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Version,
			pq.Array(&post.Tags),
			&post.User.Username,
			&post.CommentsCount,
		)
		if err != nil {
			return nil, err
		}

		feed = append(feed, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return feed, nil
}
