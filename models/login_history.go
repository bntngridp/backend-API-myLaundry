package models

import (
    "time"

    "gorm.io/gorm"
)

// LoginHistory records user login attempts and metadata
type LoginHistory struct {
    gorm.Model
    UserID    uint      `json:"user_id" gorm:"index"`
    Role      string    `json:"role"`
    IP        string    `json:"ip"`
    UserAgent string    `json:"user_agent"`
    Success   bool      `json:"success"`
    LoggedAt  time.Time `json:"logged_at"`
}

