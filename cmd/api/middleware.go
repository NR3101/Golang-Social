package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/NR3101/social/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

// checkPostOwnership is an authorization middleware that checks if the user is the owner of a post.
func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getUserFromContext(r)
		post := app.getPostFromContext(r)

		// check if the user is the owner of the post
		if user.ID == post.UserID {
			next.ServeHTTP(w, r)
			return
		}

		// check if the user has the required role to access the resource
		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenError(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checkRolePrecedence checks if the user has the required role to access the resource.
func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

// AuthTokenMiddleware is a middleware function that implements token-based authentication.
func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedError(w, r, fmt.Errorf("missing Authorization header"))
			return
		}

		// parse the Authorization header
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedError(w, r, fmt.Errorf("invalid Authorization header format"))
			return
		}

		// validate the token
		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		// get the user ID from the token claims
		claims, _ := jwtToken.Claims.(jwt.MapClaims)
		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		// retrieve the user from the database
		ctx := r.Context()
		user, err := app.store.Users.GetByID(ctx, fmt.Sprintf("%d", userID))
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		// set the user in the request context
		ctx = context.WithValue(ctx, "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}

// BasicAuthMiddleware is a middleware function that implements basic authentication.
func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicAuthError(w, r, fmt.Errorf("missing Authorization header"))
				return
			}

			// parse the Authorization header
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicAuthError(w, r, fmt.Errorf("invalid Authorization header format"))
				return
			}

			// decode the base64 encoded credentials
			credentials, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicAuthError(w, r, err)
				return
			}

			// check the credentials
			username := app.config.auth.basic.username
			password := app.config.auth.basic.password

			creds := strings.SplitN(string(credentials), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != password {
				app.unauthorizedBasicAuthError(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
