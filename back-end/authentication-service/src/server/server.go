package server

import (
	"authentication-service/src/config"
	"authentication-service/src/route"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

// StartServer initializes config and starts the HTTP server with graceful shutdown
func StartServer(cfg *config.Config, db *sql.DB) {
	// Get the logger from the config
	logger := cfg.Logger

	// Initialize routes with the database connection
	r := route.NewRoutes(cfg, db)

	// Set up the HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServicePort),
		Handler: r.Routes(),
	}

	// Log server initialization success
	logger.Info("✅ HTTP server setup completed successfully")

	// Run server in a goroutine
	go func() {
		logger.Info(fmt.Sprintf("✅ %s is running on port: %s", cfg.ServiceName, cfg.ServicePort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Call graceful shutdown
	GracefulShutdown(srv, 5*time.Second, logger)
}
