package customer_controller

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
)

// CreateRating submits rating for an order and systematically recalculates Courier and Branch average ratings
func CreateRating(c *gin.Context) {
	var input struct {
		OrderID      string  `json:"order_id" binding:"required"`
		CourierScore float64 `json:"courier_score" binding:"required"`
		BranchScore  float64 `json:"branch_score" binding:"required"`
		Tags         string  `json:"tags"`
		ReviewText   string  `json:"review_text"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Success: false,
			Message: "Format input rating tidak valid",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get authenticated customer ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Success: false,
			Message: "Unauthorized",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	customerID := uint(userIDVal.(float64))

	// Find the order
	var order models.Order
	if err := config.DB.Where("id = ?", input.OrderID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Success: false,
			Message: "Pesanan tidak ditemukan",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Ensure order is completed / done
	if order.Status != "done" && order.Status != "completed" {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Success: false,
			Message: "Rating hanya dapat diberikan pada pesanan yang telah selesai",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Check if already rated
	var existingRating models.Rating
	if err := config.DB.Where("order_id = ?", input.OrderID).First(&existingRating).Error; err == nil {
		c.JSON(http.StatusConflict, response.DefaultResponse{
			Success: false,
			Message: "Anda sudah memberikan rating untuk pesanan ini",
			Code:    http.StatusConflict,
		})
		return
	}

	// Default branch fallback if not set on order
	branchID := order.BranchID
	if branchID == 0 {
		branchID = 1
	}

	var courierID uint = 0
	if order.CourierID != nil {
		courierID = *order.CourierID
	}

	// Create Rating Record
	ratingRecord := models.Rating{
		OrderID:      input.OrderID,
		CustomerID:   customerID,
		CourierID:    courierID,
		BranchID:     branchID,
		CourierScore: math.Max(1.0, math.Min(5.0, input.CourierScore)),
		BranchScore:  math.Max(1.0, math.Min(5.0, input.BranchScore)),
		Tags:         input.Tags,
		ReviewText:   input.ReviewText,
	}

	if err := config.DB.Create(&ratingRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Success: false,
			Message: "Gagal menyimpan rating",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 1. SYSTEMATIC RECALCULATION: Branch Average Rating
	if branchID > 0 {
		var branchAvg struct {
			AvgScore float64
			Count    int64
		}
		config.DB.Model(&models.Rating{}).
			Where("branch_id = ?", branchID).
			Select("COALESCE(AVG(branch_score), 5.0) as avg_score, COUNT(id) as count").
			Scan(&branchAvg)

		newBranchRating := math.Round(branchAvg.AvgScore*10) / 10
		config.DB.Model(&models.Branch{}).Where("id = ?", branchID).Update("rating", newBranchRating)
	}

	// 2. SYSTEMATIC RECALCULATION: Courier Average Rating
	if courierID > 0 {
		var courierAvg struct {
			AvgScore float64
			Count    int64
		}
		config.DB.Model(&models.Rating{}).
			Where("courier_id = ?", courierID).
			Select("COALESCE(AVG(courier_score), 5.0) as avg_score, COUNT(id) as count").
			Scan(&courierAvg)

		newCourierRating := math.Round(courierAvg.AvgScore*10) / 10
		config.DB.Model(&models.User{}).Where("id = ?", courierID).Update("rating", newCourierRating)
	}

	c.JSON(http.StatusCreated, response.DefaultResponse{
		Success: true,
		Message: "Terima kasih! Ulasan & rating berhasil disimpan",
		Code:    http.StatusCreated,
		Data:    ratingRecord,
	})
}

// GetOrderRating gets rating for specific order
func GetOrderRating(c *gin.Context) {
	orderID := c.Param("order_id")
	var rating models.Rating

	if err := config.DB.Where("order_id = ?", orderID).First(&rating).Error; err != nil {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Success: false,
			Message: "Rating belum diberikan untuk pesanan ini",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Berhasil mengambil rating pesanan",
		Code:    http.StatusOK,
		Data:    rating,
	})
}

// GetBranchRatings fetches reviews for branch
func GetBranchRatings(c *gin.Context) {
	branchIDStr := c.Param("branch_id")
	branchID, _ := strconv.ParseUint(branchIDStr, 10, 32)

	var ratings []models.Rating
	config.DB.Where("branch_id = ?", branchID).Order("created_at desc").Find(&ratings)

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Berhasil mengambil ulasan cabang",
		Code:    http.StatusOK,
		Data:    ratings,
	})
}

// GetCourierRatings fetches reviews for courier
func GetCourierRatings(c *gin.Context) {
	courierIDStr := c.Param("courier_id")
	courierID, _ := strconv.ParseUint(courierIDStr, 10, 32)

	var ratings []models.Rating
	config.DB.Where("courier_id = ?", courierID).Order("created_at desc").Find(&ratings)

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Berhasil mengambil ulasan kurir",
		Code:    http.StatusOK,
		Data:    ratings,
	})
}
