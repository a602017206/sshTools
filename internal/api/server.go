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

// NewServer creates a new HTTP/WebSocket server listening on all interfaces
func NewServer(services *Services, port int) *Server {
	return NewServerWithHost(services, "", port)
}

// NewServerWithHost creates a new HTTP/WebSocket server bound to the given
// host. Passing "127.0.0.1" restricts the server to loopback, which is what
// the embedded mobile build uses so the API is never exposed to the LAN.
func NewServerWithHost(services *Services, host string, port int) *Server {
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

	addr := fmt.Sprintf(":%d", port)
	if host != "" {
		addr = fmt.Sprintf("%s:%d", host, port)
	}

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:           addr,
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
		sessHandler := handlers.NewSessionHandler(s.services.Session, s.wsHub, s.services.Connection)
		sessions.POST("/connect", sessHandler.Connect)
		sessions.POST("/:id/send", sessHandler.SendData)
		sessions.POST("/:id/resize", sessHandler.Resize)
		sessions.DELETE("/:id", sessHandler.Disconnect)
		sessions.GET("", sessHandler.ListSessions)

		// SFTP 文件操作（挂在已建立的 SSH 会话上）
		if s.services.SFTP != nil {
			sftpHandler := handlers.NewSFTPHandler(s.services.SFTP)
			sftp := sessions.Group("/:id/sftp")
			sftp.GET("/list", sftpHandler.List)
			sftp.GET("/pwd", sftpHandler.Pwd)
			sftp.GET("/stat", sftpHandler.Stat)
			sftp.POST("/mkdir", sftpHandler.Mkdir)
			sftp.POST("/rename", sftpHandler.Rename)
			sftp.DELETE("", sftpHandler.Delete)
			sftp.GET("/read", sftpHandler.ReadText)
			sftp.PUT("/write", sftpHandler.WriteText)
			sftp.POST("/upload", sftpHandler.Upload)
			sftp.GET("/download", sftpHandler.Download)
		}
	}

	// WebSocket endpoint
	api.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(s.wsHub, c.Writer, c.Request)
	})

	// TODO: Add more routes for Monitor, Settings
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
