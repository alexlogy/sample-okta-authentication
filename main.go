package main

import (
	"log/slog"
	"os"

	"sample-okta-authentication/config"
	"sample-okta-authentication/middleware"
	"sample-okta-authentication/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Create a Gin router with default middleware (logger and recovery)
	router := gin.Default()

	// Logger
	router.Use(gin.Logger())

	// Middleware
	router.Use(gin.Recovery())

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	// Initialize SAML middleware once at startup
	samlSP, err := middleware.Saml(cfg)
	if err != nil {
		slog.Error("failed to initialize SAML", "error", err)
		os.Exit(1)
	}

	// Define Routes
	routes.Routes(router, samlSP, cfg)

	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	if err := router.Run(); err != nil {
		slog.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}
