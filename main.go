package main

import (
	"device-monitor-go/api/handlers"
	"device-monitor-go/api/middleware"
	"device-monitor-go/config"
	"device-monitor-go/database"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:web/dist
var staticFiles embed.FS

func main() {
	// Load configuration
	config.LoadConfig()

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Set Gin mode
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.Default()

	// Apply middleware
	r.Use(middleware.CORS())
	r.Use(middleware.ErrorHandler())

	// API routes
	api := r.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
				"environment": config.AppConfig.Environment,
			})
		})

		// Session routes
		api.GET("/sessions", handlers.GetSessions)
		api.GET("/sessions/statistics", handlers.GetStatistics)
		api.GET("/sessions/device/:deviceId/statistics", handlers.GetDeviceStatistics)
		api.GET("/sessions/:id", handlers.GetSessionByID)
		api.GET("/sessions/:id/report", handlers.GetSessionReport)
		api.DELETE("/sessions/:id", handlers.DeleteSession)

		// Webhook routes
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/device/start", handlers.DeviceStart)
			webhooks.POST("/device/end", handlers.DeviceEnd)
			webhooks.POST("/test/start", handlers.TestWebhookStart)
			webhooks.POST("/test/end", handlers.TestWebhookEnd)
		}

		// IoT routes
		iot := api.Group("/iot")
		{
			iot.POST("/sync/:sessionId", handlers.SyncIotData)
			iot.GET("/data-points", handlers.GetIotDataPoints)
			iot.GET("/device/:deviceId/points", handlers.GetDevicePoints)
			iot.GET("/test-connection", handlers.TestIotConnection)
		}
	}

	// Static file serving
	if config.IsProduction() {
		// Production: serve embedded files
		staticFS, err := fs.Sub(staticFiles, "web/dist")
		if err != nil {
			log.Fatalf("Failed to get embedded files: %v", err)
		}

		// Serve static files with proper MIME types
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			
			// Skip API routes
			if strings.HasPrefix(path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
				return
			}

			// Remove leading slash for file system
			if strings.HasPrefix(path, "/") {
				path = path[1:]
			}

			// Default to index.html for root and non-existent paths
			if path == "" {
				path = "index.html"
			}

			// Try to open the requested file
			file, err := staticFS.Open(path)
			if err != nil {
				// For SPA, serve index.html for all non-existent paths
				indexFile, err := staticFS.Open("index.html")
				if err != nil {
					c.String(http.StatusInternalServerError, "Failed to load application")
					return
				}
				defer indexFile.Close()

				data, err := io.ReadAll(indexFile)
				if err != nil {
					c.String(http.StatusInternalServerError, "Failed to read application")
					return
				}
				
				c.Data(http.StatusOK, "text/html; charset=utf-8", data)
				return
			}
			defer file.Close()

			// Read file content
			data, err := io.ReadAll(file)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to read file")
				return
			}

			// Set proper content type based on file extension
			contentType := "application/octet-stream"
			switch {
			case strings.HasSuffix(path, ".html"):
				contentType = "text/html; charset=utf-8"
			case strings.HasSuffix(path, ".js"):
				contentType = "application/javascript; charset=utf-8"
			case strings.HasSuffix(path, ".css"):
				contentType = "text/css; charset=utf-8"
			case strings.HasSuffix(path, ".svg"):
				contentType = "image/svg+xml"
			case strings.HasSuffix(path, ".json"):
				contentType = "application/json"
			case strings.HasSuffix(path, ".png"):
				contentType = "image/png"
			case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
				contentType = "image/jpeg"
			}

			c.Data(http.StatusOK, contentType, data)
		})
	} else {
		// Development: proxy to Vite dev server
		r.NoRoute(func(c *gin.Context) {
			// Skip API routes
			if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
				return
			}

			// Proxy to Vite dev server
			remote, _ := url.Parse("http://localhost:5173")
			proxy := httputil.NewSingleHostReverseProxy(remote)
			proxy.ServeHTTP(c.Writer, c.Request)
		})
	}

	// Start server
	port := config.AppConfig.Port
	log.Printf("Server starting on port %s in %s mode", port, config.AppConfig.Environment)
	
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// formatJSON helper for pretty printing JSON in development
func formatJSON(data interface{}) string {
	if config.IsDevelopment() {
		return fmt.Sprintf("%+v", data)
	}
	return ""
}