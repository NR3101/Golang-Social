package main

import (
	"net/http"
	"time"

	"github.com/NR3101/social/internal/auth"
	"github.com/NR3101/social/internal/mailer"
	"github.com/NR3101/social/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// application struct holds the configuration and storage for the application
type application struct {
	config        config             // configuration for the application
	store         store.Storage      // storage interface for database operations
	logger        *zap.SugaredLogger // logger for logging messages
	mailer        mailer.Client      // mailer client for sending emails
	authenticator auth.Authenticator // authenticator for handling user authentication
}

// config struct holds the database configuration
type config struct {
	addr        string
	db          dbConfig
	env         string
	mail        mailConfig
	frontendURL string     // URL for the frontend application
	auth        authConfig // authentication configuration
}

// authConfig struct holds the authentication configuration
type authConfig struct {
	basic basicAuthConfig // configuration for basic authentication
	token tokenAuthConfig // configuration for token-based authentication
}

// tokenAuthConfig struct holds the token-based authentication configuration
type tokenAuthConfig struct {
	secret string        // secret key for signing tokens
	exp    time.Duration // expiration time for tokens
	iss    string        // issuer of the tokens
}

// basicAuthConfig struct holds the basic authentication configuration
type basicAuthConfig struct {
	username string // username for basic authentication
	password string // password for basic authentication
}

// mailConfig struct holds the email configuration
type mailConfig struct {
	exp       time.Duration  // expiration time for email tokens
	fromEmail string         // email address to send from
	sendGrid  sendGridConfig // configuration for SendGrid mailer
	mailTrap  mailTrapConfig // configuration for Mailtrap mailer
}

// mailtrapConfig struct holds the Mailtrap configuration
type mailTrapConfig struct {
	apiKey string // API key for Mailtrap
}

// sendGridConfig struct holds the SendGrid configuration
type sendGridConfig struct {
	apiKey string // API key for SendGrid
}

// dbConfig struct holds the database configuration
type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

// Func to create a new http handler with routes mounted
func (app *application) mount() http.Handler {
	r := chi.NewRouter() // create a new router

	// A good base middleware stack
	r.Use(middleware.RequestID) // adds a unique request ID to each request
	r.Use(middleware.RealIP)    // extracts the real IP address of the client
	r.Use(middleware.Logger)    // logs the start and end of each request
	r.Use(middleware.Recoverer) // recovers from panics and writes a 500 response

	// Set a timeout value on the request context (ctx), that will signal through ctx.Done() that the request has timed out and further processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// Group routes under /v1
	r.Route("/v1", func(r chi.Router) {
		// GET endpoint for health check with basic authentication
		r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler) // Health check endpoint

		// Routes related to posts
		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware) // Middleware to authenticate requests using token-based authentication

			r.Post("/", app.createPostHandler) // Create a new post

			r.Route("/{postID}", func(r chi.Router) {
				// Middleware to extract post ID from URL and load the post into the request context
				r.Use(app.postsContextMiddleware)

				r.Get("/", app.getPostHandler)       // Get a specific post by ID
				r.Delete("/", app.deletePostHandler) // Delete a specific post by ID
				r.Patch("/", app.updatePostHandler)  // Update a specific post by ID
			})
		})

		// Routes related to users
		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler) // Activate a user account with a token

			r.Route("/{userID}", func(r chi.Router) {
				// Middleware to authenticate requests using token-based authentication
				r.Use(app.AuthTokenMiddleware)

				r.Get("/", app.getUserHandler)              // Get a specific user by ID
				r.Put("/follow", app.followUserHandler)     // Follow a user
				r.Put("/unfollow", app.unfollowUserHandler) // Unfollow a user
			})

			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)         // Middleware to authenticate requests using token-based authentication
				r.Get("/feed", app.getUserFeedHandler) // Get the feed for the authenticated user
			})
		})

		// Routes related to authentication
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler) // Register a new user
			r.Post("/token", app.createTokenHandler) // Create a new authentication token
		})
	})

	return r
}

// Function to run the HTTP server with the given multiplexer
func (app *application) run(mux http.Handler) error {
	srv := http.Server{
		Addr:         app.config.addr,  // address to listen on
		Handler:      mux,              // HTTP request multiplexer
		WriteTimeout: time.Second * 30, // max duration before timing out writes of the response
		ReadTimeout:  time.Second * 10, // max duration for reading the entire request, including the body
		IdleTimeout:  time.Minute,      // idle connections timeout
	}

	app.logger.Infow("Server started", "addr", app.config.addr, "env", app.config.env)

	return srv.ListenAndServe()
}
