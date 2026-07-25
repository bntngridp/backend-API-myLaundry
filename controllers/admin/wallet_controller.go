package admin_controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
)

type ConfirmDepositRequest struct {
	CourierID uint    `json:"courier_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Notes     string  `json:"notes"`
}

// GetAdminFinancialSummary - Get overall financial dashboard statistics & courier wallets
func GetAdminFinancialSummary(c *gin.Context) {
	var totalRevenue float64
	var digitalRevenue float64
	var cashRevenue float64
	var cashOnHandCouriers float64
	var totalDepositedToAdmin float64

	// 1. Total Paid Revenue
	config.DB.Model(&models.Order{}).
		Where("payment_status = ?", "PAID").
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&totalRevenue)

	// 2. Digital Midtrans Revenue (QRIS, BANK_TRANSFER, E_WALLET)
	config.DB.Model(&models.Order{}).
		Where("payment_status = ? AND payment_method IN ('QRIS', 'BANK_TRANSFER', 'E_WALLET', 'MIDTRANS')", "PAID").
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&digitalRevenue)

	// 3. Cash Revenue
	config.DB.Model(&models.Order{}).
		Where("payment_status = ? AND payment_method = 'CASH'", "PAID").
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&cashRevenue)

	// 4. Cash on hand held by all couriers
	config.DB.Model(&models.CourierWallet{}).
		Select("COALESCE(SUM(cash_on_hand), 0)").
		Scan(&cashOnHandCouriers)

	// 5. Total cash deposited to Admin
	config.DB.Model(&models.CourierCashDeposit{}).
		Where("status = ?", "CONFIRMED").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalDepositedToAdmin)

	// 6. Preload courier wallets with User details
	var courierWallets []models.CourierWallet
	config.DB.Preload("Courier").Find(&courierWallets)

	// 7. Recent deposits
	var deposits []models.CourierCashDeposit
	config.DB.Preload("Courier").Preload("Admin").Order("id desc").Limit(20).Find(&deposits)

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Ringkasan keuangan berhasil diambil",
		Data: map[string]interface{}{
			"summary": map[string]interface{}{
				"total_revenue":            totalRevenue,
				"digital_midtrans_revenue": digitalRevenue,
				"cash_revenue":             cashRevenue,
				"cash_on_hand_couriers":    cashOnHandCouriers,
				"total_deposited_admin":    totalDepositedToAdmin,
			},
			"courier_wallets": courierWallets,
			"recent_deposits": deposits,
		},
	})
}

// ConfirmCourierDeposit - Admin accepts cash deposit from Courier
func ConfirmCourierDeposit(c *gin.Context) {
	var req ConfirmDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Payload request tidak valid: " + err.Error(),
		})
		return
	}

	adminIDVal, exists := c.Get("user_id")
	if !exists {
		adminIDVal, exists = c.Get("userID")
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "Pengguna tidak terautentikasi",
		})
		return
	}

	var adminID uint
	switch v := adminIDVal.(type) {
	case float64:
		adminID = uint(v)
	case uint:
		adminID = v
	case int:
		adminID = uint(v)
	case string:
		parsed, _ := strconv.Atoi(v)
		adminID = uint(parsed)
	}

	var wallet models.CourierWallet
	if err := config.DB.Where("courier_id = ?", req.CourierID).First(&wallet).Error; err != nil {
		wallet = models.CourierWallet{
			CourierID:     req.CourierID,
			CashOnHand:    0,
			TotalEarnings: 0,
		}
		config.DB.Create(&wallet)
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Nominal setoran harus lebih besar dari 0",
		})
		return
	}

	// Update Wallet Cash on Hand
	wallet.CashOnHand -= req.Amount
	if wallet.CashOnHand < 0 {
		wallet.CashOnHand = 0
	}
	config.DB.Save(&wallet)

	// Create Deposit Record
	deposit := models.CourierCashDeposit{
		CourierID:   req.CourierID,
		AdminID:     adminID,
		Amount:      req.Amount,
		Notes:       req.Notes,
		Status:      "CONFIRMED",
		DepositedAt: time.Now(),
	}
	config.DB.Create(&deposit)

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Setoran tunai kurir berhasil dikonfirmasi",
		Data: map[string]interface{}{
			"deposit_id":   deposit.ID,
			"courier_id":   req.CourierID,
			"amount":       req.Amount,
			"cash_on_hand": wallet.CashOnHand,
		},
	})
}
