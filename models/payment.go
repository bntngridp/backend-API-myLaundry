package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	OrderID        uint       `json:"order_id"`
	Order          Order      `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	PaymentType    string     `gorm:"size:50" json:"payment_type"` // CASH, QRIS, BANK_TRANSFER, E_WALLET
	SnapToken      string     `gorm:"size:255" json:"snap_token,omitempty"`
	SnapRedirectURL string    `gorm:"size:255" json:"snap_redirect_url,omitempty"`
	TransactionID  string     `gorm:"size:255;index" json:"transaction_id,omitempty"`
	GrossAmount    float64    `json:"gross_amount"`
	PaymentStatus  string     `gorm:"size:50;default:'PENDING'" json:"payment_status"` // PENDING, PAID, EXPIRED, FAILED
	PaidAt         *time.Time `json:"paid_at,omitempty"`
}
