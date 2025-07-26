package handlers

import (
	"auth-service/database"
	"auth-service/models"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = []byte(getEnv("JWT_SECRET", "your-secret-key"))

func Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Validate email format (basic validation)
	if !strings.Contains(req.Email, "@") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Validate password length
	if len(req.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existingUser models.User
	result := database.DB.Where("email = ?", req.Email).First(&existingUser)
	if result.Error == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		log.Printf("Database error: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password), bcrypt.DefaultCost,
	)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create user
	user := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if result := database.DB.Create(&user); result.Error != nil {
		log.Printf("Failed to create user: %v", result.Error)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.ID, user.Email)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Get user from database
	var user models.User
	result := database.DB.Where("email = ?", req.Email).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password), []byte(req.Password),
	); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.ID, user.Email)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProfile returns the user profile (protected endpoint example)
func GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by middleware)
	userID, ok := r.Context().Value("user_id").(float64)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	var user models.User
	result := database.DB.First(&user, uint(userID))

	if result.Error == gorm.ErrRecordNotFound {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateProfile updates user profile (protected endpoint example)
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(float64)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	var updateReq struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var user models.User
	result := database.DB.First(&user, uint(userID))
	if result.Error != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update fields
	if updateReq.Email != "" {
		user.Email = updateReq.Email
	}

	if result := database.DB.Save(&user); result.Error != nil {
		log.Printf("Failed to update user: %v", result.Error)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func generateJWT(userID uint, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
