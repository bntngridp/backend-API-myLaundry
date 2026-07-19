package admin_controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
)

func OrderComplete(c *gin.Context) {
	var body struct {
		OrderID uint `json:"order_id" form:"order_id"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Invalid input format",
			Data:    nil,
		})
		return
	}

	var order models.Order
	if err := config.DB.Preload("Service").Preload("Courier").Preload("Customer").Preload("Admin").Preload("Address").First(&order, body.OrderID).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Invalid order ID",
			Data:    nil,
		})
		return
	}

	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "User not authenticated",
			Data:    nil,
		})
		return
	}

	// Validasi apakah user memiliki role sebagai admin
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "User is not authorized as an admin",
			Data:    nil,
		})
		return
	}

	// Ubah status pesanan menjadi 'done'
	order.Status = "done"

	// Type assertion to get uint value from adminID
	adminIDUint, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "Invalid admin ID type",
			Data:    nil,
		})
		return
	}

	order.AdminID = &adminIDUint

	if err := config.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Code:    http.StatusInternalServerError,
			Success: false,
			Message: "Failed to update order status",
			Data:    nil,
		})
		return
	}

	orderResponse := response.OrderResponse{
		ID:         order.ID,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt.String(),
		UpdatedAt:  order.UpdatedAt.String(),
		TotalPrice: order.TotalPrice,
		Weight:     order.Weight,
		Quantity:   order.Quantity,
		Customer: response.UserResponse{
			ID:       order.Customer.ID,
			Username: order.Customer.Username,
			Email:    order.Customer.Email,
		},
		Admin: response.UserResponse{
			ID:       order.Admin.ID,
			Username: order.Admin.Username,
			Email:    order.Admin.Email,
		},
		Service: response.ServiceResponse{
			ID:    order.Service.ID,
			Title: order.Service.Title,
			Price: uint(order.Service.Price),
		},
		Courier: response.UserResponse{
			ID:       order.Courier.ID,
			Username: order.Courier.Username,
			Email:    order.Courier.Email,
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

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Order complete",
		Data:    orderResponse,
	})
}

func GetDashboardStats(c *gin.Context) {
	// Validasi admin role
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "User is not authorized as an admin",
			Data:    nil,
		})
		return
	}

	adminID, existsUser := c.Get("user_id")
	if !existsUser {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "User not authenticated",
			Data:    nil,
		})
		return
	}

	adminIDUint, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "Invalid admin ID type",
			Data:    nil,
		})
		return
	}

	var orders []models.Order
	if err := config.DB.Where("admin_id = ? OR admin_id IS NULL", adminIDUint).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Code:    http.StatusInternalServerError,
			Success: false,
			Message: "Failed to retrieve orders",
			Data:    nil,
		})
		return
	}

	now := time.Now()
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
	lastMonthEnd := thisMonthStart.Add(-time.Second)

	var thisMonthSales float64
	var lastMonthSales float64
	var totalSales float64
	var totalOrders int = len(orders)

	for _, order := range orders {
		if order.Status == "done" || order.Status == "completed" {
			totalSales += order.TotalPrice
			if order.CreatedAt.After(thisMonthStart) {
				thisMonthSales += order.TotalPrice
			} else if order.CreatedAt.After(lastMonthStart) && order.CreatedAt.Before(lastMonthEnd) {
				lastMonthSales += order.TotalPrice
			}
		}
	}

	var percentage float64
	var trend string = "none"

	if lastMonthSales > 0 {
		percentage = ((thisMonthSales - lastMonthSales) / lastMonthSales) * 100
		if percentage > 0 {
			trend = "up"
		} else if percentage < 0 {
			trend = "down"
		}
	} else if thisMonthSales > 0 {
		percentage = 100
		trend = "up"
	} else {
		percentage = 0
		trend = "none"
	}

	if percentage < 0 {
		percentage = -percentage
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Dashboard stats retrieved successfully",
		Data: gin.H{
			"total_sales":      totalSales,
			"sales_percentage": percentage,
			"sales_trend":      trend,
			"total_orders":     totalOrders,
		},
	})
}
