package customer_controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
)

func GetOrderDetailForCustomer(c *gin.Context) {
	customerIDStr := c.Param("customer_id")
	customerID, err := strconv.Atoi(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Success: false,
			Message: "Invalid customer ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Enforce tenant isolation for customer roles
	role, existsRole := c.Get("role")
	loggedInUserID, existsUser := c.Get("user_id")
	if existsRole && role == "customer" && existsUser {
		userIDUint, ok := loggedInUserID.(uint)
		if ok && uint(customerID) != userIDUint {
			c.JSON(http.StatusForbidden, response.DefaultResponse{
				Success: false,
				Message: "Access denied: you can only access your own orders",
				Code:    http.StatusForbidden,
			})
			return
		}
	}

	// Fetch orders by customer ID
	var orders []models.Order
	if err := config.DB.Preload("Customer").Preload("Courier").Preload("Admin").Preload("Service").Preload("Address").Where("customer_id = ?", customerID).Find(&orders).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Success: false,
			Message: "Invalid customer ID or no orders found",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if len(orders) == 0 {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Success: false,
			Message: "No orders found for this customer",
			Code:    http.StatusNotFound,
		})
		return
	}

	var orderResponses []response.OrderResponse
	for _, order := range orders {
		orderResponse := response.OrderResponse{
			ID:         order.ID,
			Status:     order.Status,
			TotalPrice: order.TotalPrice,
			Weight:     order.Weight,
			Quantity:   order.Quantity,
			CreatedAt:  order.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  order.UpdatedAt.Format("2006-01-02 15:04:05"),
			Customer: response.UserResponse{
				ID:       order.Customer.ID,
				Username: order.Customer.Username,
				Email:    order.Customer.Email,
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
			Courier: response.UserResponse{
				ID:       order.Courier.ID,
				Username: order.Courier.Username,
				Email:    order.Courier.Email,
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
		}
		orderResponses = append(orderResponses, orderResponse)
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Orders retrieved successfully",
		Code:    http.StatusOK,
		Data:    orderResponses,
	})
}

func CreateOrder(c *gin.Context) {
	var body struct {
		ServiceID uint `json:"service_id" form:"service_id"`
		AddressID uint `json:"address_id" form:"address_id"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input format"})
		return
	}

	// Ambil ID pengguna dari token JWT
	customerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not authenticated"})
		return
	}

	// Pastikan pengguna memiliki role customer
	role, exists := c.Get("role")
	if !exists || role != "customer" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User is not a customer or role not found"})
		return
	}

	// Validasi apakah alamat sudah dibuat oleh pengguna
	var address models.Address
	if body.AddressID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Please create an address first"})
		return
	} else {
		if err := config.DB.Where("id = ? AND customer_id = ?", body.AddressID, customerID.(uint)).First(&address).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid address ID or address does not belong to the logged-in user"})
			return
		}
	}

	// Ambil service dari database
	var service models.Service
	if err := config.DB.First(&service, body.ServiceID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid service ID"})
		return
	}

	// Calculate total price based on service price and any other applicable factors
	// totalPrice := service.Price // Adjust calculation as needed

	// Buat order baru tanpa courier_id dan weight
	adminID := uint(1)
	order := models.Order{
		CustomerID: customerID.(uint),
		AdminID:    &adminID, // Adjust as needed
		ServiceID:  body.ServiceID,
		AddressID:  body.AddressID,
		// TotalPrice: totalPrice,
		Status: "waiting for courier approval",
	}

	if err := config.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create order", "error": err.Error()})
		return
	}

	// Preload entitas terkait sebelum mengirimkan respons
	if err := config.DB.Preload("Address").Preload("Customer").Preload("Admin").Preload("Service").First(&order, order.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve created order with associated data", "error": err.Error()})
		return
	}

	// Prepare response
	orderResponse := response.OrderResponse{
		ID:        order.ID,
		Status:    order.Status,
		CreatedAt: order.CreatedAt.String(),
		UpdatedAt: order.UpdatedAt.String(),
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

	// Return response
	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Order created successfully",
		Code:    http.StatusOK,
		Data:    orderResponse,
	})
}
