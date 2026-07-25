package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
)

type LocationUpdateRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Heading   float64 `json:"heading"`
	Speed     float64 `json:"speed"`
}

// UpdateCourierLocation - Endpoint for Courier device to update real-time GPS location
func UpdateCourierLocation(c *gin.Context) {
	var req LocationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Payload lokasi tidak valid: " + err.Error(),
		})
		return
	}

	courierIDVal, exists := c.Get("user_id")
	if !exists {
		courierIDVal, exists = c.Get("userID")
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "Pengguna tidak terautentikasi",
		})
		return
	}

	var courierID uint
	switch v := courierIDVal.(type) {
	case float64:
		courierID = uint(v)
	case uint:
		courierID = v
	case int:
		courierID = uint(v)
	case string:
		parsed, _ := strconv.Atoi(v)
		courierID = uint(parsed)
	}

	var location models.CourierLocation
	err := config.DB.Where("courier_id = ?", courierID).First(&location).Error
	now := time.Now()
	if err != nil {
		location = models.CourierLocation{
			CourierID: courierID,
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Heading:   req.Heading,
			Speed:     req.Speed,
			UpdatedAt: now,
		}
		config.DB.Create(&location)
	} else {
		location.Latitude = req.Latitude
		location.Longitude = req.Longitude
		location.Heading = req.Heading
		location.Speed = req.Speed
		location.UpdatedAt = now
		config.DB.Save(&location)
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Lokasi kurir berhasil diperbarui",
		Data:    location,
	})
}

// Haversine formula to calculate distance in meters between two lat/lng coordinates
func calculateHaversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// GetOrderLiveTracking - Get real-time courier position, route polyline, distance remaining & ETA
func GetOrderLiveTracking(c *gin.Context) {
	orderIDParam := c.Param("id")
	orderID, err := strconv.Atoi(orderIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "ID Pesanan tidak valid",
		})
		return
	}

	var order models.Order
	if err := config.DB.Preload("Customer").Preload("Courier").Preload("Address").First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Code:    http.StatusNotFound,
			Success: false,
			Message: "Pesanan tidak ditemukan",
		})
		return
	}

	// Default Outlet location (Bandung)
	outletLat := -6.917464
	outletLng := 107.619123

	// Customer destination location
	destLat := -6.921464
	destLng := 107.625123
	destAddressStr := "Alamat Pemesan"

	if order.Address.StreetName != "" {
		destAddressStr = fmt.Sprintf("%s, %s, %s", order.Address.StreetName, order.Address.District, order.Address.City)
	}

	// Courier location
	var courierLat, courierLng, heading, speed float64
	courierName := "Kurir Pengantar"
	if order.CourierID != nil {
		courierName = order.Courier.Username
		var loc models.CourierLocation
		if err := config.DB.Where("courier_id = ?", *order.CourierID).First(&loc).Error; err == nil {
			courierLat = loc.Latitude
			courierLng = loc.Longitude
			heading = loc.Heading
			speed = loc.Speed
		}
	}

	// Fallback courier position if not set
	if courierLat == 0 && courierLng == 0 {
		courierLat = -6.919400
		courierLng = 107.622500
		heading = 45.0
		speed = 25.0
	}

	// Calculate remaining distance in meters & km
	distMeters := calculateHaversineMeters(courierLat, courierLng, destLat, destLng)
	distKM := distMeters / 1000.0

	var distFormatted string
	if distMeters < 1000 {
		distFormatted = fmt.Sprintf("%.0f meter", distMeters)
	} else {
		distFormatted = fmt.Sprintf("%.1f km", distKM)
	}

	// Calculate ETA in minutes (Average motorcycle speed 25 km/h = ~416 meters per minute)
	etaMinutes := int(math.Ceil(distMeters / 350.0))
	if etaMinutes < 1 {
		etaMinutes = 1
	}

	// Generate polyline route points (Outlet -> Courier -> Destination)
	midPoint1Lat := outletLat + (courierLat-outletLat)*0.5
	midPoint1Lng := outletLng + (courierLng-outletLng)*0.5
	midPoint2Lat := courierLat + (destLat-courierLat)*0.5
	midPoint2Lng := courierLng + (destLng-courierLng)*0.5

	routePoints := []map[string]float64{
		{"lat": outletLat, "lng": outletLng},
		{"lat": midPoint1Lat, "lng": midPoint1Lng},
		{"lat": courierLat, "lng": courierLng},
		{"lat": midPoint2Lat, "lng": midPoint2Lng},
		{"lat": destLat, "lng": destLng},
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Data pelacakan kurir live berhasil diambil",
		Data: map[string]interface{}{
			"order_id":         order.ID,
			"order_status":     order.Status,
			"courier_id":       order.CourierID,
			"courier_name":     courierName,
			"courier_phone":    order.Courier.PhoneNumber,
			"courier_location": map[string]interface{}{
				"latitude":  courierLat,
				"longitude": courierLng,
				"heading":   heading,
				"speed":     speed,
			},
			"outlet_location": map[string]interface{}{
				"name":      "Outlet myLaundry Central",
				"latitude":  outletLat,
				"longitude": outletLng,
			},
			"destination_location": map[string]interface{}{
				"address":   destAddressStr,
				"latitude":  destLat,
				"longitude": destLng,
			},
			"remaining_distance_meters": math.Round(distMeters),
			"remaining_distance_km":     math.Round(distKM*10) / 10,
			"remaining_distance_str":    distFormatted,
			"eta_minutes":               etaMinutes,
			"route_points":              routePoints,
		},
	})
}
