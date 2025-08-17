package main

import (
	"net/http"
)

// This file contains error handling functions for the application.

// internalServerError handles internal server errors and writes a JSON response.
func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

// badRequestError handles bad request errors and writes a JSON response.
func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("bad request error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

// notFoundError handles not found errors and writes a JSON response.
func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusNotFound, "the requested resource could not be found")
}

// unauthorizedError handles unauthorized errors and writes a JSON response.
func (app *application) unauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusUnauthorized, "unauthorized access to the requested resource")
}

// unauthorizedBasicAuthError handles unauthorized errors and writes a JSON response for basic authentication.
func (app *application) unauthorizedBasicAuthError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	// Set the WWW-Authenticate header to indicate that basic authentication is required so that we get a popup in the browser
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted", charset="UTF-8"`)

	writeJSONError(w, http.StatusUnauthorized, "unauthorized access to the requested resource")
}

// forbiddenError handles forbidden errors and writes a JSON response.
func (app *application) forbiddenError(w http.ResponseWriter, r *http.Request) {
	app.logger.Errorw("forbidden error", "method", r.Method, "path", r.URL.Path)

	writeJSONError(w, http.StatusForbidden, "you do not have permission to access this resource")
}

// rateLimitExceededError handles rate limit exceeded errors and writes a JSON response.
func (app *application) rateLimitExceededError(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Errorw("rate limit exceeded error", "method", r.Method, "path", r.URL.Path)

	// Set the Retry-After header to indicate when the client can retry the request
	w.Header().Set("Retry-After", retryAfter)

	writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, please try again after "+retryAfter)
}
