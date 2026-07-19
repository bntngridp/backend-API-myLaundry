package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordResetOTP struct {
	gorm.Model
	Email     string    `json:"email" gorm:"index"`
	OTP       string    `json:"otp"`
	ExpiresAt time.Time `json:"expires_at"`
}
