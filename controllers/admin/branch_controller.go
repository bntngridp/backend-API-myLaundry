package admin_controllers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
)

// calculateHaversineDistance calculates Earth distance in km between two lat/lng coordinates
func calculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in kilometers
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	rLat1 := lat1 * math.Pi / 180.0
	rLat2 := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(rLat1)*math.Cos(rLat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// GetBranches fetches all branches and dynamically calculates Haversine distance if lat & lng query params exist
func GetBranches(c *gin.Context) {
	var branches []models.Branch

	query := config.DB.Order("id asc")

	// Filter inactive for non-admin if desired
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Find(&branches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Success: false,
			Message: "Gagal mengambil daftar cabang",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Parse user lat & lng if provided
	userLatStr := c.Query("lat")
	userLngStr := c.Query("lng")

	hasUserLocation := false
	var userLat, userLng float64
	if userLatStr != "" && userLngStr != "" {
		if lat, err1 := strconv.ParseFloat(userLatStr, 64); err1 == nil {
			if lng, err2 := strconv.ParseFloat(userLngStr, 64); err2 == nil {
				userLat = lat
				userLng = lng
				hasUserLocation = true
			}
		}
	}

	var branchResponses []response.BranchResponse
	for _, b := range branches {
		distKm := 0.0
		if hasUserLocation {
			distKm = calculateHaversineDistance(userLat, userLng, b.Latitude, b.Longitude)
		}

		branchResponses = append(branchResponses, response.BranchResponse{
			ID:         b.ID,
			Name:       b.Name,
			Address:    b.Address,
			Latitude:   b.Latitude,
			Longitude:  b.Longitude,
			DistanceKm: math.Round(distKm*10) / 10, // Round to 1 decimal place e.g. 0.2
			Rating:     b.Rating,
			ImageURL:   b.ImageURL,
			IsActive:   b.IsActive,
		})
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Berhasil mengambil daftar cabang",
		Code:    http.StatusOK,
		Data:    branchResponses,
	})
}

// GetBranchByID fetches single branch by ID
func GetBranchByID(c *gin.Context) {
	id := c.Param("id")
	var branch models.Branch

	if err := config.DB.First(&branch, id).Error; err != nil {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Success: false,
			Message: "Cabang tidak ditemukan",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Berhasil mengambil detail cabang",
		Code:    http.StatusOK,
		Data:    branch,
	})
}

// CreateBranch handles adding a new branch
func CreateBranch(c *gin.Context) {
	var input struct {
		Name      string  `json:"name" binding:"required"`
		Address   string  `json:"address" binding:"required"`
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		Rating    float64 `json:"rating"`
		ImageURL  string  `json:"image_url"`
		IsActive  bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Success: false,
			Message: "Format input tidak valid",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if input.Rating <= 0 {
		input.Rating = 4.8
	}

	branch := models.Branch{
		Name:      input.Name,
		Address:   input.Address,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
		Rating:    input.Rating,
		ImageURL:  input.ImageURL,
		IsActive:  input.IsActive,
	}

	if err := config.DB.Create(&branch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Success: false,
			Message: "Gagal menyimpan cabang baru",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, response.DefaultResponse{
		Success: true,
		Message: "Cabang berhasil ditambahkan",
		Code:    http.StatusCreated,
		Data:    branch,
	})
}

// UpdateBranch handles editing branch details
func UpdateBranch(c *gin.Context) {
	id := c.Param("id")
	var branch models.Branch

	if err := config.DB.First(&branch, id).Error; err != nil {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Success: false,
			Message: "Cabang tidak ditemukan",
			Code:    http.StatusNotFound,
		})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Success: false,
			Message: "Format input tidak valid",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := config.DB.Model(&branch).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Success: false,
			Message: "Gagal memperbarui data cabang",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Cabang berhasil diperbarui",
		Code:    http.StatusOK,
		Data:    branch,
	})
}

// DeleteBranch removes a branch by ID
func DeleteBranch(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.Branch{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Success: false,
			Message: "Gagal menghapus cabang",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Cabang berhasil dihapus",
		Code:    http.StatusOK,
	})
}
