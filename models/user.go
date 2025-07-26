package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Email     string    `json:"email" gorm:"uniqueIndex;not null"`
    Password  string    `json:"-" gorm:"column:password_hash;not null"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
    return "users"
}

type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}