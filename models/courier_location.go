package models

import (
	"time"

	"gorm.io/gorm"
)

type CourierLocation struct {
	gorm.Model
	CourierID uint      `gorm:"uniqueIndex" json:"courier_id"`
	Courier   User      `json:"courier,omitempty" gorm:"foreignKey:CourierID"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Heading   float64   `json:"heading"` // Arah orientasi motor (0-360 derajat)
	Speed     float64   `json:"speed"`   // Kecepatan km/jam
	UpdatedAt time.Time `json:"updated_at"`
}
