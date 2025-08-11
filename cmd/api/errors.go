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
