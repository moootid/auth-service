package handlers

import (
	"auth-service/database"
	"auth-service/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// AnalyzeVideoProxy redirects requests to the AnalyzeVideo handler at http://localhost:8000/video/analyze
// and adds the user ID to the request body
func AnalyzeVideoProxy(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(float64)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	// Read the original request body
	var originalBody map[string]interface{}
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		r.Body.Close()

		// Parse the original JSON body if it exists
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &originalBody); err != nil {
				log.Printf("Error parsing JSON body: %v", err)
				http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
				return
			}
		} else {
			originalBody = make(map[string]interface{})
		}
	} else {
		originalBody = make(map[string]interface{})
	}

	// Add user ID to the request body
	originalBody["user"] = uint(userID)

	// Marshal the modified body
	modifiedBodyBytes, err := json.Marshal(originalBody)
	if err != nil {
		log.Printf("Error marshaling modified body: %v", err)
		http.Error(w, "Error preparing request", http.StatusInternalServerError)
		return
	}

	// Create a new request to the video analysis service
	var analyzeLink = []byte(getEnv("ANALYZE_VIDEO_URL", "http://localhost:8000"))
	analyzeURL := fmt.Sprintf("%s/analyze-video", analyzeLink)
	req, err := http.NewRequest(r.Method, analyzeURL, bytes.NewBuffer(modifiedBodyBytes))

	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Error creating request to video service", http.StatusInternalServerError)
		return
	}

	// Copy headers from the original request (except Authorization)
	for name, values := range r.Header {
		if name != "Authorization" && name != "Content-Length" {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}
	}

	// Set content type for JSON
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(modifiedBodyBytes)))

	// Make the request to the video analysis service
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request to video service: %v", err)
		http.Error(w, "Error connecting to video service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set response status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error copying response body: %v", err)
		return
	}

	log.Printf("Successfully proxied video analysis request for user %d", uint(userID))
}

// GetVideoAnalyses gets all video analysis jobs for the authenticated user
func GetVideoAnalyses(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(float64)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	// Get video analysis jobs from the database filtered by user ID
	var videoAnalyses []models.VideoAnalysis
	result := database.DB.Where("created_by = ?", uint(userID)).Order("created_at DESC").Find(&videoAnalyses)

	if result.Error != nil {
		log.Printf("Error retrieving video analyses for user %d: %v", uint(userID), result.Error)
		http.Error(w, "Error retrieving video analyses", http.StatusInternalServerError)
		return
	}

	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Return the video analyses as JSON
	if err := json.NewEncoder(w).Encode(videoAnalyses); err != nil {
		log.Printf("Error encoding video analyses response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved %d video analyses for user %d", len(videoAnalyses), uint(userID))
}

// GetVideoAnalysesInfo gets information about a specific video analysis job by ID
func GetVideoAnalysesInfo(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(float64)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	// Get the job ID from URL path
	vars := mux.Vars(r)
	jobID := vars["id"]

	// Validate UUID format
	if _, err := uuid.Parse(jobID); err != nil {
		http.Error(w, "Invalid job ID format", http.StatusBadRequest)
		return
	}

	// Get video analysis job from database
	var videoAnalysis models.VideoAnalysis
	result := database.DB.Where("job_id = ? AND created_by = ?", jobID, uint(userID)).First(&videoAnalysis)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			http.Error(w, "Video analysis not found or access denied", http.StatusNotFound)
			return
		}
		log.Printf("Error retrieving video analysis %s for user %d: %v", jobID, uint(userID), result.Error)
		http.Error(w, "Error retrieving video analysis information", http.StatusInternalServerError)
		return
	}

	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Return the video analysis as JSON
	if err := json.NewEncoder(w).Encode(videoAnalysis); err != nil {
		log.Printf("Error encoding video analysis response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved video analysis %s for user %d", jobID, uint(userID))
}
