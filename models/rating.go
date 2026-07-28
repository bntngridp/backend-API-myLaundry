package models

import "time"

type Rating struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	OrderID      string     `gorm:"type:varchar(50);not null;unique" json:"order_id"`
	CustomerID   uint       `gorm:"not null" json:"customer_id"`
	Customer     User       `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	CourierID    *uint      `json:"courier_id,omitempty"`
	Courier      User       `json:"courier,omitempty" gorm:"foreignKey:CourierID"`
	BranchID     uint       `gorm:"not null" json:"branch_id"`
	Branch       Branch     `json:"branch,omitempty" gorm:"foreignKey:BranchID"`
	CourierScore float64    `gorm:"type:double;not null" json:"courier_score"` // 1.0 - 5.0
	BranchScore  float64    `gorm:"type:double;not null" json:"branch_score"`  // 1.0 - 5.0
	Tags         string     `gorm:"type:varchar(255)" json:"tags"`             // e.g. "Kurir Cepat & Ramah, Cucian Rapi"
	ReviewText   string     `gorm:"type:text" json:"review_text"`
	AdminReply   string     `gorm:"type:text" json:"admin_reply,omitempty"`
	RepliedAt    *time.Time `json:"replied_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
