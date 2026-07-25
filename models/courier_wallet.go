package models

import (
	"time"

	"gorm.io/gorm"
)

type CourierWallet struct {
	gorm.Model
	CourierID     uint    `gorm:"uniqueIndex" json:"courier_id"`
	Courier       User    `json:"courier,omitempty" gorm:"foreignKey:CourierID"`
	CashOnHand    float64 `gorm:"default:0" json:"cash_on_hand"`    // Uang tunai COD di tangan kurir
	TotalEarnings float64 `gorm:"default:0" json:"total_earnings"`  // Total pendapatan/komisi kurir
}

type CourierCashDeposit struct {
	gorm.Model
	CourierID   uint       `json:"courier_id"`
	Courier     User       `json:"courier,omitempty" gorm:"foreignKey:CourierID"`
	AdminID     uint       `json:"admin_id"`
	Admin       User       `json:"admin,omitempty" gorm:"foreignKey:AdminID"`
	Amount      float64    `json:"amount"`
	Notes       string     `gorm:"size:255" json:"notes"`
	Status      string     `gorm:"size:50;default:'CONFIRMED'" json:"status"` // CONFIRMED, PENDING
	DepositedAt time.Time  `json:"deposited_at"`
}
