package models

import (
	"time"

	"gorm.io/gorm"
)

type ChatMessage struct {
	gorm.Model
	OrderID     uint      `json:"order_id" gorm:"index"`
	Order       Order     `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	SenderID    uint      `json:"sender_id"`
	Sender      User      `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	SenderRole  string    `json:"sender_role" gorm:"size:50"` // "courier", "customer", "admin"
	Message     string    `json:"message" gorm:"type:longtext"`
	ImageURL    string    `json:"image_url,omitempty" gorm:"type:longtext"` // Base64 or Image URL for Proof of Delivery
	MessageType string    `json:"message_type" gorm:"size:50;default:'TEXT'"` // "TEXT", "IMAGE", "VIDEO", "AUDIO", "DELIVERY_PROOF"
	SentAt      time.Time `json:"sent_at"`
}
