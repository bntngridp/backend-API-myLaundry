package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
)

type SendChatRequest struct {
	Message     string `json:"message"`
	ImageURL    string `json:"image_url"`
	MessageType string `json:"message_type"` // TEXT, DELIVERY_PROOF
}

// SendChatMessage - Send a chat message or Photo Proof of Delivery for an order
func SendChatMessage(c *gin.Context) {
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

	var req SendChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Payload chat tidak valid: " + err.Error(),
		})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		userIDVal, exists = c.Get("userID")
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "Pengguna tidak terautentikasi",
		})
		return
	}

	var senderID uint
	switch v := userIDVal.(type) {
	case float64:
		senderID = uint(v)
	case uint:
		senderID = v
	case int:
		senderID = uint(v)
	case string:
		parsed, _ := strconv.Atoi(v)
		senderID = uint(parsed)
	}

	userRoleVal, _ := c.Get("role")
	senderRole, _ := userRoleVal.(string)
	if senderRole == "" {
		senderRole = "admin"
	}

	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Code:    http.StatusNotFound,
			Success: false,
			Message: "Pesanan tidak ditemukan",
		})
		return
	}

	messageType := req.MessageType
	if messageType == "" {
		messageType = "TEXT"
	}

	msgText := req.Message
	if msgText == "" && messageType == "DELIVERY_PROOF" {
		msgText = "📷 Bukti Pengiriman: Laundry telah digantung di lokasi (pagar / depan kamar)."
	}

	chatMsg := models.ChatMessage{
		OrderID:     uint(orderID),
		SenderID:    senderID,
		SenderRole:  senderRole,
		Message:     msgText,
		ImageURL:    req.ImageURL,
		MessageType: messageType,
		SentAt:      time.Now(),
	}

	if err := config.DB.Create(&chatMsg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Code:    http.StatusInternalServerError,
			Success: false,
			Message: "Gagal menyimpan pesan chat",
		})
		return
	}

	// If courier uploaded Delivery Proof, update Order status
	if messageType == "DELIVERY_PROOF" {
		order.Status = "selesai_digantung"
		config.DB.Save(&order)
	}

	// Preload Sender info for response
	config.DB.Preload("Sender").First(&chatMsg, chatMsg.ID)

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Pesan chat / Bukti pengiriman berhasil dikirim",
		Data:    chatMsg,
	})
}

// GetOrderChatMessages - Retrieve chat history & proof of delivery for an order
func GetOrderChatMessages(c *gin.Context) {
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

	var messages []models.ChatMessage
	config.DB.Preload("Sender").Where("order_id = ?", orderID).Order("id asc").Find(&messages)

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Histori chat pesanan berhasil diambil",
		Data:    messages,
	})
}
