package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"easyflow-oauth2-server/pkg/api/middleware"
	"easyflow-oauth2-server/pkg/api/routes/admin"
	"easyflow-oauth2-server/pkg/api/routes/auth"
	"easyflow-oauth2-server/pkg/api/routes/oauth"
	"easyflow-oauth2-server/pkg/api/routes/user"
	"easyflow-oauth2-server/pkg/config"
	"easyflow-oauth2-server/pkg/database"
	"easyflow-oauth2-server/pkg/logger"
	"easyflow-oauth2-server/pkg/retry"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"database/sql"

	cors "github.com/OnlyNico43/gin-cors" // CORS middleware
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/valkey-io/valkey-go"
)

func main() {
	// Load configuration
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	log := logger.NewLogger(os.Stdout, "Main", cfg.LogLevel, "System")

	// Connect to database with retry logic
	db, err := retry.WithRetry(func() (*sql.DB, error) {
		return sql.Open("postgres", cfg.DatabaseURL)
	}, log, retry.DefaultRetryConfig("sql.Open"))()
	if err != nil {
		log.PrintfError("Failed to connect to database: %v", err)
		panic(err)
	}
	defer db.Close()

	// Creating migration driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.PrintfError("Failed to create migration driver: %v", err)
		panic(err)
	}
	defer driver.Close()

	// Get current working directory
	pwd, err := os.Getwd()
	if err != nil {
		log.PrintfError("Failed to get working directory: %v", err)
		panic(err)
	}

	// Initialize migrate instance
	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s/%s", pwd, cfg.MigrationsPath), "postgres", driver)
	if err != nil {
		log.PrintfError("Failed to initialize migrations: %v", err)
		panic(err)
	}
	defer m.Close()

	// Apply migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.PrintfError("Migration failed: %v", err)
		panic(err)
	}

	log.Printf("Database migration completed successfully")

	// Initialize database queries
	queries := database.New(db)

	log.Printf("Initialized Queries")

	// Connect to Valkey with retry logic
	valkeyClient, err := retry.WithRetry(func() (valkey.Client, error) {
		return valkey.NewClient(valkey.ClientOption{
			Username:    cfg.ValkeyUsername,
			Password:    cfg.ValkeyPassword,
			ClientName:  cfg.ValkeyClientName,
			InitAddress: []string{cfg.ValkeyURL},
		})
	}, log, retry.DefaultRetryConfig("valkey.NewClient"))()
	if err != nil {
		log.PrintfError("Failed to connect to Valkey: %v", err)
		panic(err)
	}
	defer valkeyClient.Close()

	log.Printf("Connected to Valkey")

	// Generate the ed25519 key pair for JWT signing from secret
	key := ed25519.NewKeyFromSeed([]byte(cfg.JwtSecret))

	log.Printf("Generated ed25519 key from secret")

	spkiBytes, err := x509.MarshalPKIXPublicKey(key.Public().(ed25519.PublicKey))
	if err != nil {
		log.PrintfError("Failed to marshal public key: %v", err)
		panic(err)
	}
	pemBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: spkiBytes,
	}
	pemBytes := pem.EncodeToMemory(&pemBlock)
	log.Printf("Public Key:\n%s", pemBytes)

	// Disable Gin debug logs
	gin.SetMode(gin.ReleaseMode)

	// Initialize Gin router
	router := gin.New()

	// Configure trusted proxies
	err = router.SetTrustedProxies(cfg.TrustedProxies)
	if err != nil {
		log.PrintfError("Could not set trusted proxies list: %v", err)
		panic(err)
	}

	// Configure router path handling
	router.RedirectFixedPath = true     // Redirect to the correct path if case-insensitive match found
	router.RedirectTrailingSlash = true // Automatically handle trailing slashes

	// Set up CORS middleware
	corsMiddleware := cors.CorsMiddleware(cors.Config{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})

	// Add middleware
	router.Use(middleware.ConfigMiddleware(cfg))
	router.Use(middleware.QueriesMiddleware(queries))
	router.Use(middleware.ValkeyMiddleware(valkeyClient))
	router.Use(middleware.KeyMiddlware(&key))
	router.Use(gin.Recovery())

	// Add endpoints
	adminEndpoints := router.Group("/admin")
	adminEndpoints.Use(corsMiddleware)
	log.PrintfInfo("Registering admin endpoints")
	admin.RegisterAdminEndpoints(adminEndpoints)

	authEndpoints := router.Group("/auth")
	authEndpoints.Use(corsMiddleware)
	log.PrintfInfo("Registering auth endpoints")
	auth.RegisterAuthEnpoints(authEndpoints)

	oauthEndpoints := router.Group("/oauth")
	log.PrintfInfo("Registering oauth endpoints")
	oauth.RegisterOAuthEndpoints(oauthEndpoints)

	userEndpoints := router.Group("/user")
	userEndpoints.Use(corsMiddleware)
	log.PrintfInfo("Registering user endpoints")
	user.RegisterUserEndpoints(userEndpoints)

	// Start server
	log.PrintfInfo("Starting Server on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.PrintfError("Failed to start server: %v", err)
		panic(err)
	}
}
