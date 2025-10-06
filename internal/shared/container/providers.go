// Package container provides fx modules for dependency injection
package container

import (
	"crypto/ed25519"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"easyflow-oauth2-server/pkg/config"
	"easyflow-oauth2-server/pkg/database"
	"easyflow-oauth2-server/pkg/logger"
	"easyflow-oauth2-server/pkg/retry"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	// Import file source for migrations.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/valkey-io/valkey-go"
	"go.uber.org/fx"
)

// ProvidersModule contains all the basic infrastructure providers.
var ProvidersModule = fx.Module("providers",
	fx.Provide(
		NewConfig,
		NewLoggerFactory,
		NewDatabase,
		NewQueries,
		NewValkeyClient,
		NewPrivateKey,
	),
)

// NewConfig provides the application configuration.
func NewConfig() (*config.Config, error) {
	return config.LoadDefaultConfig()
}

// NewLoggerFactory provides the logger factory instance.
func NewLoggerFactory(cfg *config.Config) *logger.Factory {
	return logger.NewLoggerFactory(os.Stdout, "Main", cfg.LogLevel)
}

// NewDatabase provides the database connection with migrations.
func NewDatabase(cfg *config.Config, loggerFactory *logger.Factory) (*sql.DB, error) {
	log := loggerFactory.NewLogger("System")

	// Run migrations with separate connection first
	if err := runMigrations(cfg, log); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Connect to database with retry logic for main connection
	db, err := retry.WithRetry(func() (*sql.DB, error) {
		return sql.Open("postgres", cfg.DatabaseURL)
	}, log, retry.DefaultRetryConfig("sql.Open"))()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("Database connection established successfully")
	return db, nil
}

// runMigrations runs database migrations using a separate connection.
func runMigrations(cfg *config.Config, log *logger.Logger) error {
	// Create separate connection for migrations
	migrationDB, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open migration database connection: %w", err)
	}
	defer func() {
		if err := migrationDB.Close(); err != nil {
			log.PrintfError("Failed to close migration database connection: %v", err)
		}
	}()

	// Creating migration driver
	driver, err := postgres.WithInstance(migrationDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Get current working directory
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Initialize migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s/%s", pwd, cfg.MigrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.PrintfWarning("Failed to close migration source: %v", srcErr)
		}
		if dbErr != nil {
			log.PrintfWarning("Failed to close migration database: %v", dbErr)
		}
	}()

	// Apply migrations
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Printf("Database migration completed successfully")
	return nil
}

// NewQueries provides database queries instance.
func NewQueries(db *sql.DB) *database.Queries {
	return database.New(db)
}

// NewValkeyClient provides Valkey client connection.
func NewValkeyClient(cfg *config.Config, loggerFactory *logger.Factory) (valkey.Client, error) {
	log := loggerFactory.NewLogger("System")

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
		return nil, fmt.Errorf("failed to connect to Valkey: %w", err)
	}

	log.Printf("Connected to Valkey successfully")
	return valkeyClient, nil
}

// NewPrivateKey generates the ed25519 private key for JWT signing.
func NewPrivateKey(cfg *config.Config, loggerFactory *logger.Factory) (*ed25519.PrivateKey, error) {
	log := loggerFactory.NewLogger("System")

	// Generate the ed25519 key pair for JWT signing from secret
	key := ed25519.NewKeyFromSeed([]byte(cfg.JwtSecret))

	log.Printf("Generated ed25519 key from secret")

	// Log the public key for debugging
	spkiBytes, err := x509.MarshalPKIXPublicKey(key.Public().(ed25519.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	pemBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: spkiBytes,
	}
	pemBytes := pem.EncodeToMemory(&pemBlock)
	log.Printf("Public Key:\n%s", pemBytes)

	return &key, nil
}
