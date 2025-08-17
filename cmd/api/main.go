package main

import (
	"expvar"
	"runtime"
	"time"

	"github.com/NR3101/social/internal/auth"
	"github.com/NR3101/social/internal/db"
	"github.com/NR3101/social/internal/env"
	"github.com/NR3101/social/internal/mailer"
	"github.com/NR3101/social/internal/rateLimiter"
	"github.com/NR3101/social/internal/store"
	"github.com/NR3101/social/internal/store/cache"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const version = ""

func main() {
	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:3000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redisCfg: redisConfig{
			addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
			password: env.GetString("REDIS_PASSWORD", ""),
			db:       env.GetInt("REDIS_DB", 0),
			enabled:  env.GetBool("REDIS_ENABLED", false),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3, // 3 days expiration for email tokens
			fromEmail: env.GetString("FROM_EMAIL", "crazyleakey4@typingsquirrel.com"),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
			mailTrap: mailTrapConfig{
				apiKey: env.GetString("MAILTRAP_API_KEY", ""),
			},
		},
		auth: authConfig{
			basic: basicAuthConfig{
				username: env.GetString("BASIC_AUTH_USERNAME", "admin"),
				password: env.GetString("BASIC_AUTH_PASSWORD", "admin"),
			},
			token: tokenAuthConfig{
				secret: env.GetString("TOKEN_SECRET", "averylongandsupersecuresecretkeythatshouldbeatleast256characterslongsothatitcanbeusedforjwt"),
				exp:    time.Hour * 24 * 3, // 3 days
				iss:    env.GetString("TOKEN_ISSUER", "SocialApp"),
			},
		},
		rateLimiter: rateLimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATE_LIMITER_REQUESTS_PER_TIME_FRAME", 20),
			TimeFrame:           time.Second * 5,
			Enabled:             env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	// logger initialization
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync() // flushes buffer, if any

	// Initialize the database connection
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()
	logger.Info("Connected to database pool successfully")

	// Redis Cache initialization
	var redisCache *redis.Client
	if cfg.redisCfg.enabled {
		redisCache = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.password, cfg.redisCfg.db)
		logger.Info("Connected to Redis cache successfully")
	}

	// Initialize the rate limiter
	fixedWindowRateLimiter := rateLimiter.NewFixedWindowRateLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	// Pass the database connection to the storage layer
	storage := store.NewStorage(db)
	cacheStorage := cache.NewRedisStorage(redisCache)

	// Initialize the mailer client
	// mailerClient := mailer.NewSendGridMailer(cfg.mail.fromEmail, cfg.mail.sendGrid.apiKey)
	mailerClient, err := mailer.NewMailTrapClient(cfg.mail.mailTrap.apiKey, cfg.mail.fromEmail)
	if err != nil {
		logger.Info("Error initializing mailer client: %v", err)
	}

	// Initialize the authenticator
	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.iss)

	// Create the application instance with the configuration and storage
	app := &application{
		config:        cfg,
		store:         storage,
		cacheStorage:  cacheStorage,
		logger:        logger,
		mailer:        mailerClient,
		authenticator: jwtAuthenticator,
		rateLimiter:   fixedWindowRateLimiter,
	}

	// Send metrics using expvar
	expvar.NewString("version").Set(version) // Register the version of the application
	// Register the database and goroutine stats
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// Mount the routes and start the server
	mux := app.mount()
	logger.Fatal(app.run(mux))
}
