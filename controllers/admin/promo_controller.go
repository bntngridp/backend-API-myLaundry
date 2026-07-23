package admin_controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
	"github.com/raihansyahrin/backend_laundry_app.git/utils"
)

// parseTimeFlexible parses date strings in common web formats
func parseTimeFlexible(timeStr string) *time.Time {
	if strings.TrimSpace(timeStr) == "" {
		return nil
	}
	timeStr = strings.TrimSpace(timeStr)
	
	formats := []string{
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, fmtStr := range formats {
		if t, err := time.Parse(fmtStr, timeStr); err == nil {
			return &t
		}
	}
	return nil
}

// GetPromos fetches all promos for admin (active & inactive), or active-only for customers
func GetPromos(c *gin.Context) {
	var promos []models.Promo

	// Check if user is admin via context or Authorization header token
	isAdmin := false
	if role, exists := c.Get("role"); exists && role.(string) == "admin" {
		isAdmin = true
	} else {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if claims, err := utils.ValidateJWT(tokenStr); err == nil && claims.Role == "admin" {
				isAdmin = true
			}
		}
	}

	query := config.DB.Order("created_at desc")

	if !isAdmin {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Find(&promos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Success: false,
			Message: "Gagal mengambil daftar promo",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Filter out expired promos for non-admin users
	now := time.Now()
	var validPromos []models.Promo
	for _, p := range promos {
		if !isAdmin {
			if p.ExpiredAt != nil && p.ExpiredAt.Before(now) {
				continue
			}
		}
		validPromos = append(validPromos, p)
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Berhasil mengambil daftar promo",
		Code:    http.StatusOK,
		Data:    validPromos,
	})
}

// ValidatePromo checks if a promo code is valid and returns promo details
func ValidatePromo(c *gin.Context) {
	code := c.Param("code")
	if strings.TrimSpace(code) == "" {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Success: false,
			Message: "Kode promo tidak boleh kosong",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var promo models.Promo
	if err := config.DB.Where("LOWER(code) = LOWER(?) AND is_active = ?", strings.TrimSpace(code), true).First(&promo).Error; err != nil {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Success: false,
			Message: "Kode promo tidak ditemukan atau sudah tidak aktif",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Check expiry
	if promo.ExpiredAt != nil && promo.ExpiredAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Success: false,
			Message: "Kode promo sudah kadaluarsa",
			Code:    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Kode promo valid!",
		Code:    http.StatusOK,
		Data:    promo,
	})
}

// CreatePromo adds a new promo (Admin only)
func CreatePromo(c *gin.Context) {
	var body struct {
		Code               string  `json:"code" form:"code" binding:"required"`
		Title              string  `json:"title" form:"title" binding:"required"`
		Subtitle           string  `json:"subtitle" form:"subtitle"`
		DiscountPercentage int     `json:"discount_percentage" form:"discount_percentage"`
		MaxDiscountAmount  float64 `json:"max_discount_amount" form:"max_discount_amount"`
		MinOrderAmount     float64 `json:"min_order_amount" form:"min_order_amount"`
		IsActive           *bool   `json:"is_active" form:"is_active"`
		ExpiredAt          string  `json:"expired_at" form:"expired_at"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Success: false,
			Message: "Input promo tidak valid",
			Code:    http.StatusBadRequest,
		})
		return
	}

	isActive := true
	if body.IsActive != nil {
		isActive = *body.IsActive
	}

	expiredAt := parseTimeFlexible(body.ExpiredAt)

	newPromo := models.Promo{
		Code:               strings.TrimSpace(body.Code),
		Title:              strings.TrimSpace(body.Title),
		Subtitle:           strings.TrimSpace(body.Subtitle),
		DiscountPercentage: body.DiscountPercentage,
		MaxDiscountAmount:  body.MaxDiscountAmount,
		MinOrderAmount:     body.MinOrderAmount,
		IsActive:           isActive,
		ExpiredAt:          expiredAt,
	}

	if err := config.DB.Create(&newPromo).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Success: false,
			Message: "Gagal membuat promo. Kode promo mungkin sudah digunakan.",
			Code:    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusCreated, response.DefaultResponse{
		Success: true,
		Message: "Promo berhasil dibuat!",
		Code:    http.StatusCreated,
		Data:    newPromo,
	})
}

// UpdatePromo updates an existing promo (Admin only)
func UpdatePromo(c *gin.Context) {
	id := c.Param("id")

	var promo models.Promo
	if err := config.DB.First(&promo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Success: false,
			Message: "Promo tidak ditemukan",
			Code:    http.StatusNotFound,
		})
		return
	}

	var body struct {
		Code               string   `json:"code" form:"code"`
		Title              string   `json:"title" form:"title"`
		Subtitle           string   `json:"subtitle" form:"subtitle"`
		DiscountPercentage *int     `json:"discount_percentage" form:"discount_percentage"`
		MaxDiscountAmount  *float64 `json:"max_discount_amount" form:"max_discount_amount"`
		MinOrderAmount     *float64 `json:"min_order_amount" form:"min_order_amount"`
		IsActive           *bool    `json:"is_active" form:"is_active"`
		ExpiredAt          string   `json:"expired_at" form:"expired_at"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Success: false,
			Message: "Input pembaharuan tidak valid",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if body.Code != "" {
		promo.Code = strings.TrimSpace(body.Code)
	}
	if body.Title != "" {
		promo.Title = strings.TrimSpace(body.Title)
	}
	if body.Subtitle != "" {
		promo.Subtitle = strings.TrimSpace(body.Subtitle)
	}
	if body.DiscountPercentage != nil {
		promo.DiscountPercentage = *body.DiscountPercentage
	}
	if body.MaxDiscountAmount != nil {
		promo.MaxDiscountAmount = *body.MaxDiscountAmount
	}
	if body.MinOrderAmount != nil {
		promo.MinOrderAmount = *body.MinOrderAmount
	}
	if body.IsActive != nil {
		promo.IsActive = *body.IsActive
	}
	if body.ExpiredAt != "" {
		promo.ExpiredAt = parseTimeFlexible(body.ExpiredAt)
	}

	if err := config.DB.Save(&promo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Success: false,
			Message: "Gagal memperbarui promo",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Promo berhasil diperbarui!",
		Code:    http.StatusOK,
		Data:    promo,
	})
}

// DeletePromo removes a promo (Admin only)
func DeletePromo(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.Promo{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Success: false,
			Message: "Gagal menghapus promo",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Promo berhasil dihapus!",
		Code:    http.StatusOK,
	})
}
