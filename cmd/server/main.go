// Package main implements the entrypoint for the OAuth2 server application.
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
