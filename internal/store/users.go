package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Custom errors for user-related operations
var (
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrDuplicateUsername = errors.New("username already exists")
)

// User represents a user in the system.
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	// Password is a custom type that handles password hashing and storage.
	Password  password `json:"-"` // Password should not be returned in responses
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	IsActive  bool     `json:"is_active"` // Indicates if the user account is active
}

type password struct {
	text *string
	hash []byte
}

// Set hashes the password and sets it to the User struct.
func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

// UserStore implements the Storage interface for user-related operations.
type UserStore struct {
	db *sql.DB
}

// Create inserts a new user into the database.
func (u *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3)
			RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, user.Username, user.Email, user.Password.hash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

// delete removes a user from the database by their ID.
func (u *UserStore) delete(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows // User not found
		}
		return err
	}
	return nil
}

// Delete removes a user from the database by their ID.
func (u *UserStore) Delete(ctx context.Context, userID int64) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		// delete the user
		if err := u.delete(ctx, tx, userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}

		// delete the user invitation if it exists
		if err := u.deleteUserInvitation(ctx, tx, userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}

		return nil
	})
}

// createUserInvitation creates a new user invitation in the database.
func (u *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	query := `INSERT INTO user_invitations (token, user_id,expiry) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a user by their ID from the database.
func (u *UserStore) GetByID(ctx context.Context, userID string) (*User, error) {
	query := `SELECT id, username, email, created_at, updated_at FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := u.db.QueryRowContext(ctx, query, userID).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

// CreateAndInvite creates a new user and sends an invitation email (not implemented here).
func (u *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		// Create the user in the database
		if err := u.Create(ctx, tx, user); err != nil {
			return err
		}

		// Create an invitation
		if err := u.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}

		return nil
	})
}

// getUserFromInvitation retrieves a user based on the invitation token.
func (u *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `SELECT u.id, u.username, u.email, u.created_at, u.updated_at,u.is_active
			FROM users u
			JOIN user_invitations ui ON u.id = ui.user_id
			WHERE ui.token = $1 AND ui.expiry > $2`

	// hash the plain token so it matches the stored hash, and we can compare it securely
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

// updateUser updates the user in the database.
func (u *UserStore) updateUser(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username = $1, email = $2, updated_at = NOW(), is_active = $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

// deleteUserInvitation deletes a user invitation
func (u *UserStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

// Activate activates a user account using a token.
func (u *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		// Get the user associated with the token
		user, err := u.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return ErrNotFound
			}
			return err
		}

		// Update the user's active status
		user.IsActive = true
		if err := u.updateUser(ctx, tx, user); err != nil {
			return err
		}

		// Delete the user invitation
		if err := u.deleteUserInvitation(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}
