package main

import (
	"net/http"

	"github.com/NR3101/social/internal/store"
)

// getUserFeedHandler handles requests to retrieve the user feed for the current user.
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedFeedQuery{
		Limit:  20,     // Default limit for pagination
		Offset: 0,      // Default offset for pagination
		Sort:   "desc", // Default sort order
	}

	// Parse pagination parameters from the request
	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Validate the pagination parameters
	if err := Validate.Struct(fq); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(137), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeJSONResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
