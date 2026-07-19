package models

import (
	"gorm.io/gorm"
)

type Service struct {
	gorm.Model
	Title    string  `json:"title" form:"title"`
	Time     int     `json:"time" form:"time"`
	Price    float64 `json:"price" form:"price"`
	Category string  `json:"category" form:"category"`
	AdminID  *uint   `json:"admin_id"`
}
