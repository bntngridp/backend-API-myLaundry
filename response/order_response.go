package response

type OrderResponse struct {
	ID         uint            `json:"id"`
	Status     string          `json:"status"`
	CreatedAt  string          `json:"created_at"`
	UpdatedAt  string          `json:"updated_at"`
	TotalPrice float64         `json:"total_price,omitempty"`
	Weight     float64         `json:"weight,omitempty"`
	Quantity   int             `json:"quantity,omitempty"`
	BranchID   uint            `json:"branch_id,omitempty"`
	Branch     BranchResponse  `json:"branch"`
	Customer   UserResponse    `json:"customer"`
	Courier    UserResponse    `json:"courier"`
	Admin      UserResponse    `json:"admin"`
	Service    ServiceResponse `json:"service"`
	Address    AddressResponse `json:"address"`
}
