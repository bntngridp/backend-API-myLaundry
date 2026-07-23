package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username         string    `json:"username"`
	Email            string    `json:"email" gorm:"unique"`
	PhoneNumber      string    `json:"phone_number" gorm:"unique"`
	Password         string    `json:"password"`
	Role             string    `json:"role"` // "customer", "admin", "courier"
	CreatedByAdminID *uint     `json:"created_by_admin_id"`
	CreatedByAdmin    *User     `json:"created_by_admin" gorm:"foreignKey:CreatedByAdminID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Addresses        []Address    `gorm:"foreignkey:CustomerID"`
	LoginHistories   []LoginHistory `gorm:"foreignKey:UserID"`
}
