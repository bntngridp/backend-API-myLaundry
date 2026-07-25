package response

type BranchResponse struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	DistanceKm float64 `json:"distance_km"`
	Rating     float64 `json:"rating"`
	ImageURL   string  `json:"image_url"`
	IsActive   bool    `json:"is_active"`
}
