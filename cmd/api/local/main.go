package main

import (
	"log"
	"os"
	api "stori-challenge/cmd/api"
)

func main() {
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		os.Setenv("GIN_MODE", "debug")
	}

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	server := api.NewServer()
	if err := server.Run("0.0.0.0:" + httpPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}


