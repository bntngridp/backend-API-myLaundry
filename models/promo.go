package models

import "time"

type Promo struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	Code               string     `gorm:"type:varchar(50);unique;not null" json:"code"`
	Title              string     `gorm:"type:varchar(100);not null" json:"title"`
	Subtitle           string     `gorm:"type:varchar(255)" json:"subtitle"`
	DiscountPercentage int        `gorm:"default:0" json:"discount_percentage"`
	MaxDiscountAmount  float64    `gorm:"default:0" json:"max_discount_amount"`
	MinOrderAmount     float64    `gorm:"default:0" json:"min_order_amount"`
	IsActive           bool       `gorm:"default:true" json:"is_active"`
	ExpiredAt          *time.Time `json:"expired_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
