package main

import (
	"context"
	"net/http"

	"github.com/NR3101/social/internal/store"
	"github.com/go-chi/chi/v5"
)

// CreatePostPayload represents the payload for creating a new post
type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,min=1,max=255"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

// UpdatePostPayload represents the payload for updating an existing post
type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,min=1,max=255"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

// createPostHandler handles the creation of a new post.
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	// Create a new post
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Validate the payload
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		// TODO : change after authentication is implemented
		UserID: 1,
	}

	ctx := r.Context()
	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeJSONResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getPostHandler handles the retrieval of a specific post by ID.
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := app.getPostFromContext(r)

	// Retrieve comments for the post
	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	if err := app.writeJSONResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// deletePostHandler handles the deletion of a specific post by ID.
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "postID")

	ctx := r.Context()
	if err := app.store.Posts.Delete(ctx, postID); err != nil {
		if err == store.ErrNotFound {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	app.writeJSONResponse(w, http.StatusOK, map[string]string{
		"postID":  postID,
		"message": "Post deleted successfully",
	})
}

// updatePostHandler handles the update of a specific post by ID.
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the post from the context
	post := app.getPostFromContext(r)

	// Marshal the request body into the UpdatePostPayload
	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Validate the payload
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Update the post fields if they are provided in the payload
	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}

	// Update the post in the store
	ctx := r.Context()
	if err := app.store.Posts.Update(ctx, post); err != nil {
		if err == store.ErrNotFound {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeJSONResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// postsContextMiddleware is a middleware that retrieves a post by its ID from the URL
func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := chi.URLParam(r, "postID")

		ctx := r.Context()
		post, err := app.store.Posts.GetByID(ctx, postID)
		if err != nil {
			if err == store.ErrNotFound {
				app.notFoundError(w, r, err)
				return
			}
			app.internalServerError(w, r, err)
			return
		}

		// Store the post in the context for use in subsequent handlers
		ctx = context.WithValue(ctx, "post", post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getPostFromContext retrieves the post from the request context.
func (app *application) getPostFromContext(r *http.Request) *store.Post {
	post, _ := r.Context().Value("post").(*store.Post)

	return post
}
