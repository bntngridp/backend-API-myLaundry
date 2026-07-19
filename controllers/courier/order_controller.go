package courier_controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
)

func AcceptOrder(c *gin.Context) {
	orderID := c.Param("id")

	var body struct {
		CourierID uint `json:"courier_id" form:"courier_id"`
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
	if err := config.DB.Preload("Service").Preload("Courier").Preload("Customer").Preload("Admin").Preload("Address").First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Invalid order ID",
			Data:    nil,
		})
		return
	}

	// Cek apakah order sudah diterima sebelumnya
	if order.CourierID != nil {
		c.JSON(http.StatusConflict, response.DefaultResponse{
			Code:    http.StatusConflict,
			Success: false,
			Message: "Order has already been accepted by another courier",
			Data:    nil,
		})
		return
	}

	courierID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "User not authenticated",
			Data:    nil,
		})
		return
	}

	// Validasi apakah user memiliki role sebagai kurir
	userRole, exists := c.Get("role")
	if !exists || userRole != "courier" {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "User is not authorized as a courier",
			Data:    nil,
		})
		return
	}

	// Konversi courierID ke uint
	courierIDUint, ok := courierID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "Invalid courier ID type",
			Data:    nil,
		})
		return
	}

	// Tetapkan courier_id dan ubah status order
	order.CourierID = &courierIDUint // Gunakan courierIDUint yang telah dikonversi
	order.Status = "Kurir On The Way"
	order.AdminID = nil // Hapus pengaturan AdminID karena tidak ada admin yang menerima order

	if err := config.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Code:    http.StatusInternalServerError,
			Success: false,
			Message: "Failed to accept order",
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
		Message: "Order accepted successfully",
		Data:    orderResponse,
	})
}

func CourierArrived(c *gin.Context) {
	var body struct {
		OrderID    uint    `json:"order_id" form:"order_id"`
		Weight     float64 `json:"weight,omitempty" form:"weight"`
		Quantity   int     `json:"quantity,omitempty" form:"quantity"`
		TotalPrice float64 `json:"total_price,omitempty" form:"total_price"`
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

	// Fetch the service to get its price
	var service models.Service
	if err := config.DB.First(&service, order.ServiceID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve service details"})
		return
	}

	var totalPrice float64
	if body.TotalPrice > 0 {
		totalPrice = body.TotalPrice
		order.Weight = body.Weight
		order.Quantity = body.Quantity
	} else if service.Category == "Laundry Satuan" {
		order.Quantity = body.Quantity
		totalPrice = float64(service.Price) * float64(order.Quantity)
		order.Weight = 0 // Reset weight if it was set
	} else {
		order.Weight = body.Weight
		totalPrice = float64(service.Price) * order.Weight
		order.Quantity = 0 // Reset quantity if it was set
	}

	order.TotalPrice = totalPrice
	order.Status = "arrived - proses pembayaran"

	// Update the order in the database
	if err := config.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update order status and weight/quantity", "error": err.Error()})
		return
	}

	// Preload associated data before responding
	if err := config.DB.Preload("Address").Preload("Customer").Preload("Admin").Preload("Service").Preload("Courier").First(&order, order.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve updated order with associated data", "error": err.Error()})
		return
	}

	// Prepare response
	orderResponse := response.OrderResponse{
		ID:         order.ID,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt.String(),
		UpdatedAt:  order.UpdatedAt.String(),
		TotalPrice: totalPrice,
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

	if service.Title == "Laundry Satuan" {
		orderResponse.Quantity = order.Quantity
		orderResponse.Weight = 0
	} else {
		orderResponse.Weight = order.Weight
		orderResponse.Quantity = 0
	}

	// Return response
	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Courier arrived, weight/quantity updated successfully",
		Code:    http.StatusOK,
		Data:    orderResponse,
	})
}

func AcceptCashPayment(c *gin.Context) {
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
	// Pastikan untuk preload kolom yang diperlukan
	if err := config.DB.Preload("Service").Preload("Courier").Preload("Customer").Preload("Address").First(&order, body.OrderID).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Invalid order ID",
			Data:    nil,
		})
		return
	}

	// Cek apakah kurir adalah kurir yang menangani pesanan
	courierID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "User not authenticated",
			Data:    nil,
		})
		return
	}

	// Validasi apakah user memiliki role sebagai kurir
	userRole, exists := c.Get("role")
	if !exists || userRole != "courier" {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "User is not authorized as a courier",
			Data:    nil,
		})
		return
	}

	// Validasi apakah pesanan sudah tiba di lokasi
	if order.Status != "arrived - proses pembayaran" {
		c.JSON(http.StatusConflict, response.DefaultResponse{
			Code:    http.StatusConflict,
			Success: false,
			Message: "Order is not in 'arrived - proses pembayaran' status",
			Data:    nil,
		})
		return
	}

	// Konversi courierID ke uint
	courierIDUint, ok := courierID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "Invalid courier ID type",
			Data:    nil,
		})
		return
	}

	// Cek apakah kurir yang saat ini melakukan pembayaran adalah kurir yang menangani pesanan
	if order.CourierID == nil || *order.CourierID != courierIDUint {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "Courier is not authorized to accept cash payment for this order",
			Data:    nil,
		})
		return
	}

	// Hitung total harga
	var totalPrice float64
	if order.Service.Category == "Laundry Satuan" {
		totalPrice = float64(order.Service.Price) * float64(order.Quantity)
		order.TotalPrice = totalPrice
		order.Weight = 0 // Reset weight if it was set
	} else {
		totalPrice = float64(order.Service.Price) * float64(order.Weight)
		order.TotalPrice = totalPrice
		order.Quantity = 0 // Reset quantity if it was set
	}

	// Ubah status pesanan menjadi 'in progress'
	order.Status = "in progress"

	// Save order without overwriting total_price and admin_id
	if err := config.DB.Model(&order).Updates(map[string]interface{}{
		"status":      order.Status,
		"total_price": order.TotalPrice,
	}).Error; err != nil {
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
		Courier: response.UserResponse{
			ID:       order.Courier.ID,
			Username: order.Courier.Username,
			Email:    order.Courier.Email,
		},
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

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Cash payment accepted, order in progress",
		Data:    orderResponse,
	})
}

func OrderDelivery(c *gin.Context) {
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

	courierID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "User not authenticated",
			Data:    nil,
		})
		return
	}

	// Validasi apakah user memiliki role sebagai kurir
	userRole, exists := c.Get("role")
	if !exists || userRole != "courier" {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "User is not authorized as a courier",
			Data:    nil,
		})
		return
	}

	// Validasi apakah pesanan sudah selesai di proses
	if order.Status != "done" {
		c.JSON(http.StatusConflict, response.DefaultResponse{
			Code:    http.StatusConflict,
			Success: false,
			Message: "Order is not marked as done",
			Data:    nil,
		})
		return
	}

	// Konversi courierID ke uint
	courierIDUint, ok := courierID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Code:    http.StatusUnauthorized,
			Success: false,
			Message: "Invalid courier ID type",
			Data:    nil,
		})
		return
	}

	// Ubah status pesanan menjadi 'delivering' dan set courier yang akan mengantar
	order.Status = "delivering"
	order.CourierID = &courierIDUint

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
		Message: "Order is being delivered",
		Data:    orderResponse,
	})
}
