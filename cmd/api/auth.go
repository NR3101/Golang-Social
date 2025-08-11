package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/NR3101/social/internal/mailer"
	"github.com/google/uuid"

	"github.com/NR3101/social/internal/store"
)

// RegisterUserPayload defines the structure for user registration payload
type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

// UserWithToken extends the User model with a token for activation or invitation
type UserWithToken struct {
	*store.User
	Token string `json:"token"` // Token for user activation or invitation
}

// registerUserHandler handles user registration requests
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Validate the payload
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Create a new user instance
	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	// Hash the password and set it in the user struct
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Create the user in the database
	ctx := r.Context()

	plainToken := uuid.New().String() // Generate a new UUID token for invitation

	// hash the token for security
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	if err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp); err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestError(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	userWithToken := &UserWithToken{
		User:  user,
		Token: plainToken, // Include the plain token in the response
	}

	// email variables
	isProdEnv := app.config.env == "production"
	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	// Send the welcome email using the mailer client
	_, err := app.mailer.Send(
		mailer.UserWelcomeTemplate,
		user.Username,
		user.Email,
		vars,
		!isProdEnv,
	)
	if err != nil {
		app.logger.Errorw("failed to send welcome email", "error", err, "email", user.Email)

		// rollback user creation if email sending fails(SAGA pattern)
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("failed to rollback user creation after email failure", "error", err, "userID", user.ID)
		}

		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeJSONResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
