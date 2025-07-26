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
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// TranscodeVideoProxy redirects requests to the TranscodeVideo handler at http://localhost:4000/video/transcode
// and adds the user ID to the request body
func TranscodeVideoProxy(w http.ResponseWriter, r *http.Request) {
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
	originalBody["created_by"] = uint(userID)

	// Marshal the modified body
	modifiedBodyBytes, err := json.Marshal(originalBody)
	if err != nil {
		log.Printf("Error marshaling modified body: %v", err)
		http.Error(w, "Error preparing request", http.StatusInternalServerError)
		return
	}

	// Create a new request to the video transcode service
	var transcodeLink = []byte(getEnv("TRANSCODE_VIDEO_URL", "http://localhost:4000"))
	transcodeURL := fmt.Sprintf("%s/transcode", transcodeLink)
	req, err := http.NewRequest(r.Method, transcodeURL, bytes.NewBuffer(modifiedBodyBytes))

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

	// Make the request to the video transcode service
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

	log.Printf("Successfully proxied video transcode request for user %d", uint(userID))
}

func GetVideoTranscodes(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(float64)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	// Get transcoding jobs from the database filtered by user ID
	var transcodingJobs []models.TranscodingJob
	result := database.DB.Where("created_by = ?", uint(userID)).Order("inserted_at DESC").Find(&transcodingJobs)

	if result.Error != nil {
		log.Printf("Error retrieving transcoding jobs for user %d: %v", uint(userID), result.Error)
		http.Error(w, "Error retrieving transcoding jobs", http.StatusInternalServerError)
		return
	}

	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Return the transcoding jobs as JSON
	if err := json.NewEncoder(w).Encode(transcodingJobs); err != nil {
		log.Printf("Error encoding transcoding jobs response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved %d transcoding jobs for user %d", len(transcodingJobs), uint(userID))
}

// GetVideoTranscodeInfo gets information about a specific transcoding job by ID
func GetVideoTranscodeInfo(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(float64)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	// Get the video ID from URL path
	vars := mux.Vars(r)
	videoID := vars["id"]

	// Validate UUID format
	if _, err := uuid.Parse(videoID); err != nil {
		http.Error(w, "Invalid video ID format", http.StatusBadRequest)
		return
	}

	// Get transcoding job from database
	var transcodingJob models.TranscodingJob
	result := database.DB.Where("id = ? AND created_by = ?", videoID, uint(userID)).First(&transcodingJob)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			http.Error(w, "Video not found or access denied", http.StatusNotFound)
			return
		}
		log.Printf("Error retrieving transcoding job %s for user %d: %v", videoID, uint(userID), result.Error)
		http.Error(w, "Error retrieving video information", http.StatusInternalServerError)
		return
	}

	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Return the transcoding job as JSON
	if err := json.NewEncoder(w).Encode(transcodingJob); err != nil {
		log.Printf("Error encoding transcoding job response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved transcoding job %s for user %d", videoID, uint(userID))
}

// DownloadVideoFromS3 downloads a video file from S3 and streams it to the client
func DownloadVideoFromS3(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(float64)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	// Get the video ID from URL path
	vars := mux.Vars(r)
	videoID := vars["id"]

	// Validate UUID format
	if _, err := uuid.Parse(videoID); err != nil {
		http.Error(w, "Invalid video ID format", http.StatusBadRequest)
		return
	}

	// Get transcoding job from database
	var transcodingJob models.TranscodingJob
	result := database.DB.Where("id = ? AND created_by = ?", videoID, uint(userID)).First(&transcodingJob)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			http.Error(w, "Video not found or access denied", http.StatusNotFound)
			return
		}
		log.Printf("Error retrieving transcoding job %s for user %d: %v", videoID, uint(userID), result.Error)
		http.Error(w, "Error retrieving video information", http.StatusInternalServerError)
		return
	}

	// Check if the transcoding job has an output URL (completed job)
	if transcodingJob.OutputURL == nil || *transcodingJob.OutputURL == "" {
		http.Error(w, "Video is not ready for download", http.StatusNotFound)
		return
	}

	// Initialize AWS session
	awsRegion := getEnv("AWS_REGION", "us-east-1")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(
			getEnv("AWS_ACCESS_KEY_ID", ""),
			getEnv("AWS_SECRET_ACCESS_KEY", ""),
			"",
		),
	})

	if err != nil {
		log.Printf("Error creating AWS session: %v", err)
		http.Error(w, "Error connecting to storage service", http.StatusInternalServerError)
		return
	}

	// Create S3 service client
	svc := s3.New(sess)

	// Parse the S3 URL to get bucket and key
	outputURL := *transcodingJob.OutputURL
	bucket, key, err := parseS3URL(outputURL)
	if err != nil {
		log.Printf("Error parsing S3 URL %s: %v", outputURL, err)
		http.Error(w, "Invalid video storage location", http.StatusInternalServerError)
		return
	}

	// Get object from S3
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result_s3, err := svc.GetObject(input)
	if err != nil {
		log.Printf("Error getting object from S3: %v", err)
		http.Error(w, "Error retrieving video file", http.StatusInternalServerError)
		return
	}
	defer result_s3.Body.Close()

	// Set appropriate headers for video download
	filename := filepath.Base(key)
	if filename == "" || filename == "." {
		filename = fmt.Sprintf("video_%s.mp4", videoID)
	}

	// Set content type based on file extension
	contentType := getContentType(filename)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Set content length if available
	if result_s3.ContentLength != nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", *result_s3.ContentLength))
	}

	// Stream the file to the client
	bytesWritten, err := io.Copy(w, result_s3.Body)
	if err != nil {
		log.Printf("Error streaming video file to client: %v", err)
		return
	}

	log.Printf("Successfully downloaded video %s for user %d (%d bytes)", videoID, uint(userID), bytesWritten)
}

// parseS3URL parses an S3 URL and returns bucket and key
func parseS3URL(s3URL string) (bucket, key string, err error) {
	// Remove s3:// prefix if present
	s3URL = strings.TrimPrefix(s3URL, "s3://")

	// Split into bucket and key
	parts := strings.SplitN(s3URL, "/", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid S3 URL format")
	}

	return parts[0], parts[1], nil
}

// getContentType returns the appropriate content type based on file extension
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".flv":
		return "video/x-flv"
	case ".webm":
		return "video/webm"
	case ".mkv":
		return "video/x-matroska"
	default:
		return "application/octet-stream"
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
