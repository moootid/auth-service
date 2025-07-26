package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
