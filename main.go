package main

import (
	"auth-service/database"
	"auth-service/handlers"
	"auth-service/middleware"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize database
	database.InitDB()

	// Get underlying sql.DB to properly close connection
	sqlDB, err := database.DB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// Setup routes
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"auth-service"}`))
	}).Methods("GET")

	// Public routes
	router.HandleFunc("/auth/register", handlers.Register).Methods("POST")
	router.HandleFunc("/auth/login", handlers.Login).Methods("POST")

	// Protected routes (require authentication)
	router.HandleFunc("/auth/profile",
		middleware.AuthMiddleware(handlers.GetProfile)).Methods("GET")
	router.HandleFunc("/auth/profile",
		middleware.AuthMiddleware(handlers.UpdateProfile)).Methods("PUT")
	// Video analysis routes
	router.HandleFunc("/auth/video/analyze",
		middleware.AuthMiddleware(handlers.AnalyzeVideoProxy)).Methods("POST")
	router.HandleFunc("/auth/video/analyze",
		middleware.AuthMiddleware(handlers.GetAnalyzeVideoProxy)).Methods("GET")
	// Video transcoding routes
	router.HandleFunc("/auth/video/transcode",
		middleware.AuthMiddleware(handlers.TranscodeVideoProxy)).Methods("POST")
	// Get list of video transcodes
	router.HandleFunc("/auth/video/transcode",
		middleware.AuthMiddleware(handlers.GetVideoTranscodes)).Methods("GET")
	// Get specific video transcode info
	router.HandleFunc("/auth/video/transcode/{id}",
		middleware.AuthMiddleware(handlers.GetVideoTranscodeInfo)).Methods("GET")
	// Download video from S3
	router.HandleFunc("/auth/video/transcode/{id}/download",
		middleware.AuthMiddleware(handlers.DownloadVideoFromS3)).Methods("GET")
	// CORS middleware for development
	router.Use(corsMiddleware)

	// Get port from environment
	port := getEnv("PORT", "8080")

	log.Printf("Auth service starting on port %s", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, router))
}

// CORS middleware for development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
