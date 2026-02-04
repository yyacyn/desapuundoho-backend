package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// Response structure for JSON responses
type Response struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// CORS middleware to allow frontend to access backend
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Hello endpoint handler
func helloHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message: "This was a triumph",
		Status:  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
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

	response := map[string]interface{}{
		"message": "Backend is running!",
		"status":  "healthy",
		"database": map[string]string{
			"status": dbStatus,
			"error":  dbError,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// User endpoint example
func userHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message: "User data retrieved successfully!",
		Status:  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Initialize database connection
	log.Println("Initializing database connection...")
	if err := InitDB(); err != nil {
		log.Printf("⚠️  Database connection failed: %v", err)
		log.Println("Server will start without database")
	}
	defer CloseDB()

	// Register routes
	http.HandleFunc("/api/hello", enableCORS(helloHandler))
	http.HandleFunc("/api/health", enableCORS(healthHandler))
	http.HandleFunc("/api/user", enableCORS(userHandler))

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // default port
	}

	// Format port with colon
	portWithColon := ":" + port

	log.Printf("Server starting on port %s", portWithColon)
	log.Printf("Endpoints available:")
	log.Printf("  - http://localhost:%s/api/hello", port)
	log.Printf("  - http://localhost:%s/api/health", port)

	if err := http.ListenAndServe(portWithColon, nil); err != nil {
		log.Fatal(err)
	}
}
