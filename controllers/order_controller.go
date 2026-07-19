package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
)

func UpdateOrderStatus(c *gin.Context) {
	var body struct {
		OrderID uint   `json:"order_id" form:"order_id"`
		Status  string `json:"status" form:"status"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input format"})
		return
	}

	var order models.Order
	if err := config.DB.First(&order, body.OrderID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid order ID"})
		return
	}

	// Check if the status update is valid
	if body.Status == "in progress" && order.Status != "arrived" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid status update"})
		return
	}

	order.Status = body.Status

	if err := config.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully"})
}

func GetOrders(c *gin.Context) {
	var orders []models.Order
	if err := config.DB.Preload("Customer").Preload("Courier").Preload("Admin").Preload("Service").Preload("Address").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Success: false,
			Message: "Failed to retrieve orders",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	var orderResponses []response.OrderResponse
	for _, order := range orders {
		orderResponse := response.OrderResponse{
			ID:         order.ID,
			Status:     order.Status,
			CreatedAt:  order.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  order.UpdatedAt.Format("2006-01-02 15:04:05"),
			TotalPrice: order.TotalPrice,
			Weight:     order.Weight,
			Quantity:   order.Quantity,
			Customer: response.UserResponse{
				ID:        order.Customer.ID,
				Username:  order.Customer.Username,
				Email:     order.Customer.Email,
				Role:      nil,
				Addresses: nil,
			},
			Courier: response.UserResponse{
				ID:        order.Courier.ID,
				Username:  order.Courier.Username,
				Email:     order.Courier.Email,
				Role:      nil,
				Addresses: nil},
			Admin: response.UserResponse{
				ID:        order.Admin.ID,
				Username:  order.Admin.Username,
				Email:     order.Admin.Email,
				Role:      nil,
				Addresses: nil},
			Service: response.ServiceResponse{
				ID:    order.Service.ID,
				Title: order.Service.Title,
				Price: uint(order.Service.Price),
			},
			Address: response.AddressResponse{
				ID:            order.Address.ID,
				CustomerID:    order.Address.CustomerID,
				ReceiverName:  order.Address.ReceiverName,
				PhoneNumber:   order.Address.PhoneNumber,
				HouseNumber:   order.Address.HouseNumber,
				ResidenceName: order.Address.ResidenceName,
				AddressNotes:  order.Address.AddressNotes,
				StreetName:    order.Address.StreetName,
				District:      order.Address.District,
				SubDistrict:   order.Address.SubDistrict,
				City:          order.Address.City,
				Area:          order.Address.Area,
			},
		}
		orderResponses = append(orderResponses, orderResponse)
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Successfully retrieved orders",
		Code:    http.StatusOK,
		Data:    orderResponses,
	})
}

func DeleteOrder(c *gin.Context) {
	var body struct {
		OrderID uint `json:"order_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input format"})
		return
	}

	if err := config.DB.Delete(&models.Order{}, body.OrderID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}
