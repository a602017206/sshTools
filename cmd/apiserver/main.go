package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sshTools/internal/api"
	"sshTools/internal/config"
	"sshTools/internal/service"
	"sshTools/internal/ssh"
	"sshTools/internal/store"
)

func main() {
	fmt.Println("=== sshTools API Server ===")

	// Initialize configuration manager
	configManager, err := config.NewConfigManager()
	if err != nil {
		log.Fatalf("Failed to initialize config manager: %v\n", err)
	}
	fmt.Println("✓ Configuration manager initialized")

	// Initialize credential store
	credentialStore := store.NewCredentialStore()
	fmt.Println("✓ Credential store initialized")

	// Initialize managers
	sessionManager := ssh.NewSessionManager()
	transferManager := ssh.NewTransferManager()
	fmt.Println("✓ SSH managers initialized")

	// Initialize services
	services := &api.Services{
		Connection: service.NewConnectionService(configManager, credentialStore),
		Session:    service.NewSessionService(sessionManager),
		SFTP:       service.NewSFTPService(sessionManager, transferManager),
		Monitor:    service.NewMonitorService(sessionManager),
		Settings:   service.NewSettingsService(configManager),
	}
	fmt.Println("✓ Business services initialized")

	// Create API server (default port 8080)
	port := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", &port)
	}

	server := api.NewServer(services, port)
	fmt.Println("✓ HTTP/WebSocket server created")

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v\n", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nReceived shutdown signal...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server exited successfully")
}
