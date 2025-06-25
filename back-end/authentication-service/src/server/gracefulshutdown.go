package server

import (
	"authentication-service/src/logger"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	// adjust if your logger path differs
)

// GracefulShutdown handles OS signals and shuts down the server gracefully
func GracefulShutdown(srv *http.Server, timeout time.Duration, logger *logger.Logger) {
	// Create a channel to listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Log that graceful shutdown handling is set up
	logger.Info("✅ Graceful shutdown monitoring initialized successfully")

	// Wait for signal
	<-stop
	logger.Info("🔄 Shutdown signal received. Cleaning up...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("⚠️ Graceful shutdown failed: " + err.Error())
	} else {
		logger.Info("✅ Server shut down gracefully")
	}
}
