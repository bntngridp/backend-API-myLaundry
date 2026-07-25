package models

import "time"

type Branch struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Address   string    `gorm:"type:text;not null" json:"address"`
	Latitude  float64   `gorm:"type:double;not null" json:"latitude"`
	Longitude float64   `gorm:"type:double;not null" json:"longitude"`
	Rating    float64   `gorm:"type:double;default:4.8" json:"rating"`
	ImageURL  string    `gorm:"type:text" json:"image_url"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
