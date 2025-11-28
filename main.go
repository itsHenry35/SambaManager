package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsHenry35/SambaManager/api/handlers"
	"github.com/itsHenry35/SambaManager/api/routes"
	"github.com/itsHenry35/SambaManager/config"
	"github.com/itsHenry35/SambaManager/queue"
	"github.com/itsHenry35/SambaManager/services"
)

//go:embed all:frontend/dist
var staticFiles embed.FS

// getStaticFS returns the embedded static file system
func getStaticFS() (fs.FS, error) {
	staticFS, err := fs.Sub(staticFiles, "frontend/dist")
	if err != nil {
		return nil, err
	}
	return staticFS, nil
}

func main() {
	gin.SetMode("release")

	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	if err := config.Load(*configPath); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize queue for request handling
	taskQueue := queue.NewQueue(1) // 1 worker thread

	// Initialize services
	sambaService := services.NewSambaService()

	// Initialize handlers (all using the same queue and service for thread safety)
	userHandler := handlers.NewUserHandler(sambaService, taskQueue)
	shareHandler := handlers.NewShareHandler(sambaService, taskQueue)
	userShareHandler := handlers.NewUserShareHandler(sambaService, taskQueue)
	userProfileHandler := handlers.NewUserProfileHandler(sambaService, taskQueue)
	systemHandler := handlers.NewSystemHandler()

	// Get embedded static file system
	staticFS, err := getStaticFS()
	if err != nil {
		log.Fatalf("Failed to get static files filesystem: %v", err)
	}

	// Setup Gin router
	router := gin.Default()

	// Setup routes (all handlers share the same queue to prevent concurrent smb.conf access)
	routes.SetupRoutes(router, userHandler, shareHandler, userShareHandler, userProfileHandler, systemHandler)

	// Serve embedded frontend
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Remove leading slash for fs.Stat
		fsPath := path
		if len(fsPath) > 0 && fsPath[0] == '/' {
			fsPath = fsPath[1:]
		}

		// Check if file exists in embedded FS
		if _, err := fs.Stat(staticFS, fsPath); err == nil {
			c.FileFromFS(path, http.FS(staticFS))
			return
		}
		// Serve index.html for client-side routing
		c.FileFromFS("/", http.FS(staticFS))
	})

	// Create HTTP server
	addr := config.AppConfig.Server.Host + ":" + config.AppConfig.Server.Port
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Shutdown task queue
	taskQueue.Shutdown()

	// Shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
