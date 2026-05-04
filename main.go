package main

import (
	"log"
	"os"
	"time"

	_ "desapuundoho-backend/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Puundoho API
// @version         1.0
// @description     API server for Desa Puundoho Information System.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host
// @BasePath  /api

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @description                 Paste your token with "Bearer " prefix (e.g. "Bearer <token>")

func main() {
	// Initialize database connection
	log.Println("Initializing database connection...")
	if err := InitDB(); err != nil {
		log.Printf("⚠️  Database connection failed: %v", err)
		log.Println("Server will start without database")
	}
	defer CloseDB()

	// Initialize auth (JWT secret + seed admin user)
	initAuth()

	// Set Gin mode (release for production)
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Root path to stop 404 logs
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Desa Puundoho API is running",
			"docs":    "/swagger/index.html",
		})
	})

	// Swagger route (now at /swagger/index.html)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Public routes
	api := router.Group("/api")
	{
		api.GET("/hello", helloHandler)
		api.GET("/health", healthHandler)
		api.POST("/auth/login", loginHandler)

		// Public article read endpoints
		api.GET("/articles", listArticlesHandler)
		api.GET("/articles/:id", getArticleHandler)

		// Public listings read endpoints
		api.GET("/listings", listListingsHandler)
		api.GET("/listings/:id", getListingHandler)

		// Public gallery read endpoints
		api.GET("/galeri", listGalleryHandler)
		api.GET("/galeri/:id", getGalleryHandler)

		// ImageKit auth for client uploads (public because pengaduan uploads are unauthenticated)
		api.GET("/imagekit/auth", imagekitAuthHandler)

		// Public Dusun boundary read endpoint
		api.GET("/dusun", listDusunHandler)

		// Public Resident Data (for graphs & village profile)
		api.GET("/penduduk/datasets", listDatasetsHandler)
		api.GET("/penduduk/datasets/:id/records", listPendudukByDatasetHandler)
		api.GET("/penduduk/datasets/:id/stats", getPendudukStatsHandler)

		// Public SDG Data from Kemendesa
		api.GET("/sdgs", GetLiveSDGS)
		api.GET("/idm", GetLiveIDM)

		// Public reads for APBDes and Produk
		api.GET("/produk-desa", listProdukDesaHandler)
		api.GET("/bansos", listBansosHandler)
		api.GET("/bansos/:id", getBansosHandler)
		api.GET("/stunting", listStuntingHandler)
		api.GET("/stunting/:id", getStuntingHandler)
		api.GET("/apbdes", listApbdesHandler)
		api.GET("/apbdes/:id/pendapatan", listPendapatanHandler)
		api.GET("/apbdes/:id/pengeluaran", listPengeluaranHandler)

		// Public Pengaduan endpoints
		api.GET("/pengaduan", listPengaduanHandler)
		api.GET("/pengaduan/:id", getPengaduanHandler)
		api.POST("/pengaduan", createPengaduanHandler)

		// Public Pengajuan endpoints
		api.GET("/pengajuan", listPengajuanHandler)
		api.GET("/pengajuan/:id", getPengajuanHandler)
		api.POST("/pengajuan", createPengajuanHandler)
	}

	// Protected routes (require JWT)
	protected := api.Group("")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/auth/me", meHandler)

		// Admin-only routes
		adminOnly := protected.Group("")
		adminOnly.Use(RoleMiddleware("admin"))
		{
			// Articles
			adminOnly.POST("/articles", createArticleHandler)
			adminOnly.PUT("/articles/:id", updateArticleHandler)
			adminOnly.DELETE("/articles/:id", deleteArticleHandler)

			// Listings
			adminOnly.POST("/listings", createListingHandler)
			adminOnly.PUT("/listings/:id", updateListingHandler)
			adminOnly.DELETE("/listings/:id", deleteListingHandler)

			// Gallery
			adminOnly.POST("/galeri", createGalleryHandler)
			adminOnly.PUT("/galeri/:id", updateGalleryHandler)
			adminOnly.DELETE("/galeri/:id", deleteGalleryHandler)

			// Dusun Boundaries
			adminOnly.POST("/dusun", createDusunHandler)
			adminOnly.PUT("/dusun/:id", updateDusunHandler)
			adminOnly.DELETE("/dusun/:id", deleteDusunHandler)

			// Pengaduan (Admin Only - Update Status)
			adminOnly.PATCH("/pengaduan/:id/status", updatePengaduanStatusHandler)

			// Pengajuan (Admin Only - Update Status)
			adminOnly.PATCH("/pengajuan/:id/status", updatePengajuanStatusHandler)

			// Data Penduduk (Admin Only)
			adminOnly.POST("/penduduk/datasets", createDatasetHandler)
			adminOnly.DELETE("/penduduk/datasets/:id", deleteDatasetHandler)
			adminOnly.PATCH("/penduduk/records/:id", patchPendudukHandler)
			adminOnly.POST("/penduduk/datasets/:id/bulk", bulkCreatePendudukHandler)
			adminOnly.DELETE("/penduduk/records/:id", deleteRecordHandler)

			// Stunting (Admin Only)
			adminOnly.POST("/stunting", createStuntingHandler)
			adminOnly.PUT("/stunting/:id", updateStuntingHandler)
			adminOnly.DELETE("/stunting/:id", deleteStuntingHandler)
		}

		// ImageKit auth (server-side signing for secure uploads)
		// protected.GET("/imagekit/auth", imagekitAuthHandler)

		// PDF Parser (for APBDes import)
		protected.POST("/apbdes/parse-pdf", parsePDFHandler)

		// Bendahara routes
		bendahara := protected.Group("")
		bendahara.Use(RoleMiddleware("bendahara"))
		{
			// Produk Desa
			bendahara.POST("/produk-desa", createProdukDesaHandler)
			bendahara.PUT("/produk-desa/:id", updateProdukDesaHandler)
			bendahara.DELETE("/produk-desa/:id", deleteProdukDesaHandler)

			// Bansos
			bendahara.POST("/bansos", createBansosHandler)
			bendahara.PUT("/bansos/:id", updateBansosHandler)
			bendahara.DELETE("/bansos/:id", deleteBansosHandler)

			// APBDes
			bendahara.POST("/apbdes", createApbdesHandler)
			bendahara.DELETE("/apbdes/:id", deleteApbdesHandler)

			// APBDes Pendapatan
			bendahara.POST("/apbdes/pendapatan", createPendapatanHandler)
			bendahara.PUT("/apbdes/pendapatan/:id", updatePendapatanHandler)
			bendahara.DELETE("/apbdes/pendapatan/:id", deletePendapatanHandler)

			// APBDes Pengeluaran
			bendahara.POST("/apbdes/pengeluaran", createPengeluaranHandler)
			bendahara.PUT("/apbdes/pengeluaran/:id", updatePengeluaranHandler)
			bendahara.DELETE("/apbdes/pengeluaran/:id", deletePengeluaranHandler)
		}
	}

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Start server
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📍 Public endpoints:")
	log.Printf("   - POST http://localhost:%s/api/auth/login", port)
	log.Printf("   - GET  http://localhost:%s/api/articles", port)
	log.Printf("   - GET  http://localhost:%s/api/listings", port)
	log.Printf("   - GET  http://localhost:%s/api/galeri", port)
	log.Printf("📍 Protected endpoints (JWT required):")
	log.Printf("   - GET  http://localhost:%s/api/auth/me", port)
	log.Printf("   - POST/PUT/DELETE http://localhost:%s/api/articles", port)
	log.Printf("   - POST/PUT/DELETE http://localhost:%s/api/listings", port)
	log.Printf("   - POST/PUT/DELETE http://localhost:%s/api/galeri", port)
	log.Printf("   - GET  http://localhost:%s/api/imagekit/auth", port)

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
