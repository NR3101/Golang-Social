package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// Validate is a global validator instance used for validating structs.
var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

// writeJSON writes a JSON response to the http.ResponseWriter.
func writeJSON(w http.ResponseWriter, status int, data any) error {
	// Set the content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the status code
	w.WriteHeader(status)

	// Encode the data as JSON and write it to the response
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return err
	}

	return nil
}

// readJSON reads a JSON request body and decodes it into the provided interface.
func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_576 // 1 MB
	// Set the maximum size of the request body to prevent large payloads
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Decode the JSON request body into the provided data interface
	decoder := json.NewDecoder(r.Body)
	// Disallow unknown fields to prevent unexpected data from being processed
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

// writeJSONError writes a JSON error response to the http.ResponseWriter.
func writeJSONError(w http.ResponseWriter, status int, message string) error {
	// Consistent error response structure
	type Envelope struct {
		Error string `json:"error"`
	}

	return writeJSON(w, status, &Envelope{
		Error: message,
	})
}

// writeJSONResponse writes a JSON response with a data envelope to the http.ResponseWriter.
func (app *application) writeJSONResponse(w http.ResponseWriter, status int, data any) error {
	type Envelope struct {
		Data any `json:"data"`
	}

	return writeJSON(w, status, &Envelope{
		Data: data,
	})
}
