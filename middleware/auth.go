package middleware

import (
    "context"
    "net/http"
    "os"
    "strings"

    "github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(getEnv("JWT_SECRET", "your-secret-key"))

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }
        
        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
            http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
            return
        }
        
        token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })
        
        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            return
        }
        
        // Add user info to context
        ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
        ctx = context.WithValue(ctx, "email", claims["email"])
        
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}