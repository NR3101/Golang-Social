package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/NR3101/social/internal/store"
	"github.com/go-chi/chi/v5"
)

// getUserHandler handles the retrieval of a specific user by ID.
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get the user from the request context
	user := app.getUserFromContext(r)

	// Write the user as a JSON response
	if err := app.writeJSONResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// followUserHandler handles the following of a user by another user.
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	toFollowUser := app.getUserFromContext(r) // ID of the user to be followed
	currentUserID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Followers.Follow(ctx, toFollowUser.ID, currentUserID); err != nil {
		if errors.Is(err, store.ErrAlreadyFollowing) {
			app.badRequestError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeJSONResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// unfollowUserHandler handles the unfollowing of a user by another user.
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	toUnfollowUser := app.getUserFromContext(r) // ID of the user to be unfollowed
	currentUserID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Followers.Unfollow(ctx, toUnfollowUser.ID, currentUserID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeJSONResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// activateUserHandler handles the activation of a user account using a token.
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the activation token from the URL parameters
	token := chi.URLParam(r, "token")

	// Activate the user account using the token
	ctx := r.Context()
	if err := app.store.Users.Activate(ctx, token); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	// Respond with a 204 No Content status
	if err := app.writeJSONResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// usersContextMiddleware extracts the user ID from the URL and loads the user into the request context.
func (app *application) usersContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the user ID from the URL parameters
		userID := chi.URLParam(r, "userID")

		// Retrieve the user from the store by ID
		ctx := r.Context()
		user, err := app.store.Users.GetByID(ctx, userID)
		if err != nil {
			if err == store.ErrNotFound {
				app.notFoundError(w, r, err)
				return
			}
			app.internalServerError(w, r, err)
			return
		}

		// Store the user in the request context
		ctx = context.WithValue(ctx, "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getUserFromContext retrieves the user from the request context.
func (app *application) getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value("user").(*store.User)

	return user
}
