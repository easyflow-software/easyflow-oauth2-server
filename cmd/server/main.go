// Package main implements the entrypoint for the OAuth2 server application.
//
// @title Easyflow OAuth2 Server API
// @version 1.0
// @description OAuth2 server implementation following RFC 6749 with early adaptation for OAuth2.1
// @termsOfService http://swagger.io/terms/
//
// @contact.name API Support
// @contact.email support@easyflow.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /
//
// @securityDefinitions.apikey SessionToken
// @in cookie
// @name session_token
// @description Session token for authenticated users
//
// @securityDefinitions.basic BasicAuth
// @description Basic authentication for OAuth2 client credentials
package main

import (
	"context"
	"easyflow-oauth2-server/internal/server/config"
	"easyflow-oauth2-server/internal/server/container"
	"easyflow-oauth2-server/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize basic logger for main function
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		panic(err)
	}
	mainLogger := logger.NewLogger(os.Stdout, "Main", cfg.LogLevel, "System")

	// Log PID as first log entry
	mainLogger.PrintfInfo("Starting application with PID: %d", os.Getpid())

	// Create fx application
	app := container.NewApp()

	// Handle shutdown signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start the application
	startCtx := context.Background()
	if err := app.Start(startCtx); err != nil {
		mainLogger.PrintfError("Failed to start application: %v", err)
		os.Exit(1)
	}

	// Wait for shutdown signal
	mainLogger.PrintfInfo("Application started successfully. Press Ctrl+C to shutdown...")
	<-c
	mainLogger.PrintfInfo("Shutting down server...")

	// Stop the application gracefully with timeout
	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		mainLogger.PrintfError("Failed to stop application: %v", err)
		os.Exit(1)
	}

	mainLogger.PrintfInfo("Server stopped")
}
