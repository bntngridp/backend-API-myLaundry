package controllers

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
)

type CreatePaymentRequest struct {
	OrderID     uint   `json:"order_id" binding:"required"`
	PaymentType string `json:"payment_type"` // QRIS, BANK_TRANSFER, CASH
}

type ConfirmCashRequest struct {
	OrderID uint `json:"order_id" binding:"required"`
}

// CreateSnapTransaction - Generate Midtrans Snap token/redirect URL for digital payment
func CreateSnapTransaction(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Payload request tidak valid: " + err.Error(),
		})
		return
	}

	var order models.Order
	if err := config.DB.Preload("Customer").Preload("Service").First(&order, req.OrderID).Error; err != nil {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Code:    http.StatusNotFound,
			Success: false,
			Message: "Pesanan tidak ditemukan",
		})
		return
	}

	paymentType := req.PaymentType
	if paymentType == "" {
		paymentType = "QRIS"
	}

	// Double check if payment already paid
	if order.PaymentStatus == "PAID" {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Pesanan ini sudah dibayar sebelumnya",
		})
		return
	}

	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	isProduction := os.Getenv("MIDTRANS_IS_PRODUCTION") == "true"

	transactionID := fmt.Sprintf("ORDER-%d-%d", order.ID, time.Now().Unix())
	grossAmount := order.TotalPrice
	if grossAmount <= 0 {
		grossAmount = 25000 // Fallback nominal jika belum di-set
	}

	var snapToken, snapRedirectURL string

	if serverKey != "" && serverKey != "SB-Mid-server-YOUR_SERVER_KEY" {
		// Midtrans REST API Integration via Resty
		snapURL := "https://app.sandbox.midtrans.com/snap/v1/transactions"
		if isProduction {
			snapURL = "https://app.midtrans.com/snap/v1/transactions"
		}

		client := resty.New()
		resp, err := client.R().
			SetBasicAuth(serverKey, "").
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetBody(map[string]interface{}{
				"transaction_details": map[string]interface{}{
					"order_id":     transactionID,
					"gross_amount": int64(grossAmount),
				},
				"customer_details": map[string]interface{}{
					"first_name": order.Customer.Username,
					"email":      order.Customer.Email,
					"phone":      order.Customer.PhoneNumber,
				},
				"item_details": []map[string]interface{}{
					{
						"id":       fmt.Sprintf("SVC-%d", order.ServiceID),
						"price":    int64(grossAmount),
						"quantity": 1,
						"name":     fmt.Sprintf("Service Laundry Order #%d", order.ID),
					},
				},
			}).
			Post(snapURL)

		if err == nil && resp.StatusCode() == 201 {
			var result map[string]interface{}
			if err := resp.Result(); err == nil {
				if token, ok := result["token"].(string); ok {
					snapToken = token
				}
				if redirectURL, ok := result["redirect_url"].(string); ok {
					snapRedirectURL = redirectURL
				}
			}
		}
	}

	// If Midtrans API not configured or fallback required
	if snapToken == "" {
		snapToken = fmt.Sprintf("SNAP-SIMULATED-%s", transactionID)
		snapRedirectURL = fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v2/vtweb/%s", snapToken)
	}

	// Save or update Payment record
	var payment models.Payment
	err := config.DB.Where("order_id = ?", order.ID).First(&payment).Error
	if err != nil {
		payment = models.Payment{
			OrderID:         order.ID,
			PaymentType:     paymentType,
			SnapToken:       snapToken,
			SnapRedirectURL: snapRedirectURL,
			TransactionID:   transactionID,
			GrossAmount:     grossAmount,
			PaymentStatus:   "PENDING",
		}
		config.DB.Create(&payment)
	} else {
		payment.SnapToken = snapToken
		payment.SnapRedirectURL = snapRedirectURL
		payment.TransactionID = transactionID
		payment.PaymentType = paymentType
		payment.GrossAmount = grossAmount
		payment.PaymentStatus = "PENDING"
		config.DB.Save(&payment)
	}

	// Update order payment method
	order.PaymentMethod = paymentType
	config.DB.Save(&order)

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Snap Transaction berhasil dibuat",
		Data: map[string]interface{}{
			"order_id":          order.ID,
			"snap_token":        snapToken,
			"snap_redirect_url": snapRedirectURL,
			"transaction_id":   transactionID,
			"gross_amount":     grossAmount,
			"payment_type":     paymentType,
		},
	})
}

// HandleMidtransNotification - Webhook Callback endpoint for Midtrans Realtime Verification
func HandleMidtransNotification(c *gin.Context) {
	var notification map[string]interface{}
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Payload webhook tidak valid",
		})
		return
	}

	orderIDStr, _ := notification["order_id"].(string)
	transactionStatus, _ := notification["transaction_status"].(string)
	fraudStatus, _ := notification["fraud_status"].(string)
	statusCode, _ := notification["status_code"].(string)
	grossAmountStr, _ := notification["gross_amount"].(string)
	signatureKey, _ := notification["signature_key"].(string)

	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")

	// Validate Signature Key if serverKey is available
	if serverKey != "" && signatureKey != "" {
		hash := sha512.New()
		hash.Write([]byte(orderIDStr + statusCode + grossAmountStr + serverKey))
		expectedSignature := hex.EncodeToString(hash.Sum(nil))
		if expectedSignature != signatureKey {
			c.JSON(http.StatusUnauthorized, response.DefaultResponse{
				Code:    http.StatusUnauthorized,
				Success: false,
				Message: "Signature Key tidak valid",
			})
			return
		}
	}

	// Find payment by TransactionID or OrderID
	var payment models.Payment
	if err := config.DB.Where("transaction_id = ?", orderIDStr).First(&payment).Error; err != nil {
		var realOrderID uint
		if _, err := fmt.Sscanf(orderIDStr, "ORDER-%d-", &realOrderID); err == nil {
			config.DB.Where("order_id = ?", realOrderID).First(&payment)
		}
	}

	if payment.ID == 0 {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Code:    http.StatusNotFound,
			Success: false,
			Message: "Transaksi pembayaran tidak ditemukan",
		})
		return
	}

	var newPaymentStatus string
	now := time.Now()

	if transactionStatus == "capture" {
		if fraudStatus == "challenge" {
			newPaymentStatus = "PENDING"
		} else if fraudStatus == "accept" {
			newPaymentStatus = "PAID"
			payment.PaidAt = &now
		}
	} else if transactionStatus == "settlement" {
		newPaymentStatus = "PAID"
		payment.PaidAt = &now
	} else if transactionStatus == "cancel" || transactionStatus == "deny" || transactionStatus == "expire" {
		newPaymentStatus = "FAILED"
	} else if transactionStatus == "pending" {
		newPaymentStatus = "PENDING"
	}

	payment.PaymentStatus = newPaymentStatus
	config.DB.Save(&payment)

	// Update associated Order status
	var order models.Order
	if err := config.DB.First(&order, payment.OrderID).Error; err == nil {
		if newPaymentStatus == "PAID" {
			order.PaymentStatus = "PAID"
			if order.Status == "menunggu_pembayaran" || order.Status == "pending" {
				order.Status = "sedang_diproses"
			}
			config.DB.Save(&order)
		}
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Notification berhasil diproses",
		Data: map[string]interface{}{
			"payment_status": newPaymentStatus,
			"order_id":       payment.OrderID,
		},
	})
}

// ConfirmCashPayment - Courier endpoint to confirm cash received from customer
func ConfirmCashPayment(c *gin.Context) {
	var req ConfirmCashRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Payload request tidak valid: " + err.Error(),
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

	var order models.Order
	if err := config.DB.First(&order, req.OrderID).Error; err != nil {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Code:    http.StatusNotFound,
			Success: false,
			Message: "Pesanan tidak ditemukan",
		})
		return
	}

	if order.PaymentStatus == "PAID" {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Pesanan ini sudah dikonfirmasi lunas sebelumnya",
		})
		return
	}

	now := time.Now()
	order.PaymentMethod = "CASH"
	order.PaymentStatus = "PAID"
	config.DB.Save(&order)

	// Save or Update Payment Record
	var payment models.Payment
	err := config.DB.Where("order_id = ?", order.ID).First(&payment).Error
	if err != nil {
		payment = models.Payment{
			OrderID:       order.ID,
			PaymentType:   "CASH",
			TransactionID: fmt.Sprintf("CASH-%d-%d", order.ID, now.Unix()),
			GrossAmount:   order.TotalPrice,
			PaymentStatus: "PAID",
			PaidAt:        &now,
		}
		config.DB.Create(&payment)
	} else {
		payment.PaymentType = "CASH"
		payment.PaymentStatus = "PAID"
		payment.PaidAt = &now
		config.DB.Save(&payment)
	}

	// Update Courier Wallet (Increase Cash On Hand)
	var wallet models.CourierWallet
	err = config.DB.Where("courier_id = ?", courierID).First(&wallet).Error
	if err != nil {
		wallet = models.CourierWallet{
			CourierID:     courierID,
			CashOnHand:    order.TotalPrice,
			TotalEarnings: order.TotalPrice * 0.15, // 15% komisi pengantaran kurir
		}
		config.DB.Create(&wallet)
	} else {
		wallet.CashOnHand += order.TotalPrice
		wallet.TotalEarnings += (order.TotalPrice * 0.15)
		config.DB.Save(&wallet)
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Pembayaran tunai (CASH) berhasil dikonfirmasi oleh Kurir",
		Data: map[string]interface{}{
			"order_id":       order.ID,
			"payment_status": "PAID",
			"payment_type":   "CASH",
			"cash_on_hand":   wallet.CashOnHand,
		},
	})
}
