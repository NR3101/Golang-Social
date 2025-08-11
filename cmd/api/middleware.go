package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

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
