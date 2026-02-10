package main

import (
	"log"
	"os"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	log.Println("Initializing database connection...")
	if err := InitDB(); err != nil {
		log.Printf("⚠️  Database connection failed: %v", err)
		log.Println("Server will start without database")
	}
	defer CloseDB()

	// Set Gin mode (release for production)
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// API routes
	api := router.Group("/api")
	{
		api.GET("/hello", helloHandler)
		api.GET("/health", healthHandler)
		api.GET("/user", userHandler)
	}

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Start server
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📍 Endpoints available:")
	log.Printf("   - http://localhost:%s/api/hello", port)
	log.Printf("   - http://localhost:%s/api/health", port)
	log.Printf("   - http://localhost:%s/api/user", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

// Hello handler
func helloHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "This was a triumph",
		"status":  "success",
	})
}

// Health check handler
func healthHandler(c *gin.Context) {
	dbStatus := "disconnected"
	dbError := ""

	// Check if database is connected
	if DB != nil {
		err := DB.Ping()
		if err == nil {
			dbStatus = "connected"
		} else {
			dbError = err.Error()
		}
	}

	c.JSON(200, gin.H{
		"message": "Backend is running!",
		"status":  "healthy",
		"database": gin.H{
			"status": dbStatus,
			"error":  dbError,
		},
	})
}

// User handler example
func userHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "User data retrieved successfully!",
		"status":  "success",
	})
}
