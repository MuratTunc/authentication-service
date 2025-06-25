package main

import (
	"authentication-service/src/config"
	"authentication-service/src/database"
	"authentication-service/src/server"
)

const (
	// Define the environment variable prefix as a constant
	SERVICE_PREFIX = "AUTHENTICATION"
)

func main() {
	// Load configuration from environment with the defined prefix
	cfg := config.NewConfig(SERVICE_PREFIX)

	// Get the logger from the config
	logger := cfg.Logger

	// Make Database connection with the loaded config and handle any errors
	db, err := database.ConnectToDB(cfg)
	if err != nil {
		logger.Fatalf("❌ DATABASE FATAL ERROR: Failed to connect to the database: %v", err)
	}

	// Pass db to the server
	server.StartServer(cfg, db)
}
