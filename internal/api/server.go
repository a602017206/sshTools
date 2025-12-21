package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"AHaSSHTools/internal/api/handlers"
	"AHaSSHTools/internal/api/websocket"
	"AHaSSHTools/internal/service"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP/WebSocket server
type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	wsHub      *websocket.Hub
	services   *Services
}

// Services contains all business logic services
type Services struct {
	Connection *service.ConnectionService
	Session    *service.SessionService
	SFTP       *service.SFTPService
	Monitor    *service.MonitorService
	Settings   *service.SettingsService
}

// NewServer creates a new HTTP/WebSocket server
func NewServer(services *Services, port int) *Server {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	wsHub := websocket.NewHub()

	server := &Server{
		router:   router,
		wsHub:    wsHub,
		services: services,
	}

	// Setup routes
	server.setupRoutes()

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return server
}

// setupRoutes configures all HTTP routes and WebSocket endpoints
func (s *Server) setupRoutes() {
	// Apply global middleware
	s.router.Use(CORS())
	s.router.Use(Logger())
	s.router.Use(Recovery())

	// API version 1 group
	api := s.router.Group("/api/v1")

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// Connection management routes
	connections := api.Group("/connections")
	{
		connHandler := handlers.NewConnectionHandler(s.services.Connection)
		connections.GET("", connHandler.GetConnections)
		connections.POST("", connHandler.AddConnection)
		connections.PUT("/:id", connHandler.UpdateConnection)
		connections.DELETE("/:id", connHandler.DeleteConnection)
		connections.POST("/test", connHandler.TestConnection)
	}

	// SSH session routes
	sessions := api.Group("/sessions")
	{
		sessHandler := handlers.NewSessionHandler(s.services.Session, s.wsHub)
		sessions.POST("/connect", sessHandler.Connect)
		sessions.POST("/:id/send", sessHandler.SendData)
		sessions.POST("/:id/resize", sessHandler.Resize)
		sessions.DELETE("/:id", sessHandler.Disconnect)
		sessions.GET("", sessHandler.ListSessions)
	}

	// WebSocket endpoint
	api.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(s.wsHub, c.Writer, c.Request)
	})

	// TODO: Add more routes for SFTP, Monitor, Settings
}

// Start starts the HTTP server and WebSocket hub
func (s *Server) Start() error {
	// Start WebSocket hub
	go s.wsHub.Run()

	fmt.Printf("API Server starting on %s\n", s.httpServer.Addr)
	fmt.Printf("WebSocket endpoint: ws://localhost%s/api/v1/ws\n", s.httpServer.Addr)

	// Start HTTP server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

// GetWebSocketHub returns the WebSocket hub
func (s *Server) GetWebSocketHub() *websocket.Hub {
	return s.wsHub
}
