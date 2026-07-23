package customer_controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
	"golang.org/x/crypto/bcrypt"
)

// GetCustomers retrieves all customers with their addresses
func GetCustomers(c *gin.Context) {
	role, existsRole := c.Get("role")
	loggedInUserID, existsUser := c.Get("user_id")

	query := config.DB.Preload("Addresses").Where("role = ?", "customer")

	if existsRole && existsUser {
		userIDUint, ok := loggedInUserID.(uint)
		if ok {
			roleStr, okRole := role.(string)
			if okRole {
				if roleStr == "admin" {
					query = query.Where("created_by_admin_id = ?", userIDUint)
				} else {
					var loggedInUser models.User
					if err := config.DB.First(&loggedInUser, userIDUint).Error; err == nil && loggedInUser.CreatedByAdminID != nil {
						query = query.Where("created_by_admin_id = ?", *loggedInUser.CreatedByAdminID)
					}
				}
			}
		}
	}

	var customers []models.User
	if err := query.Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve customers"})
		return
	}

	var customerResponses []response.UserResponse
	for _, customer := range customers {
		var addressResponses []response.AddressResponse
		for _, address := range customer.Addresses {
			addressResponse := response.AddressResponse{
				ID:            address.ID,
				CustomerID:    address.CustomerID,
				ReceiverName:  address.ReceiverName,
				PhoneNumber:   address.PhoneNumber,
				HouseNumber:   address.HouseNumber,
				ResidenceName: address.ResidenceName,
				AddressNotes:  address.AddressNotes,
				StreetName:    address.StreetName,
				District:      address.District,
				SubDistrict:   address.SubDistrict,
				City:          address.City,
				Area:          address.Area,
			}
			addressResponses = append(addressResponses, addressResponse)
		}

		customerResponse := response.UserResponse{
			ID:        customer.ID,
			Username:  customer.Username,
			Email:     customer.Email,
			Role:      nil,
			Addresses: &addressResponses,
		}
		customerResponses = append(customerResponses, customerResponse)
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Successfully retrieved customers",
		Code:    http.StatusOK,
		Data:    customerResponses,
	})
}

// GetCustomer retrieves a single customer with their addresses based on ID
func GetCustomer(c *gin.Context) {
	id := c.Param("id")

	var customer models.User
	if err := config.DB.Preload("Addresses").Where("role = ? AND id = ?", "customer", id).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Customer not found"})
		return
	}

	var addressResponses []response.AddressResponse
	for _, address := range customer.Addresses {
		addressResponse := response.AddressResponse{
			ID:            address.ID,
			CustomerID:    address.CustomerID,
			ReceiverName:  address.ReceiverName,
			PhoneNumber:   address.PhoneNumber,
			HouseNumber:   address.HouseNumber,
			ResidenceName: address.ResidenceName,
			AddressNotes:  address.AddressNotes,
			StreetName:    address.StreetName,
			District:      address.District,
			SubDistrict:   address.SubDistrict,
			City:          address.City,
			Area:          address.Area,
		}
		addressResponses = append(addressResponses, addressResponse)
	}

	customerResponse := response.UserResponse{
		ID:        customer.ID,
		Username:  customer.Username,
		Email:     customer.Email,
		Role:      nil,
		Addresses: &addressResponses,
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Successfully retrieved customer",
		Code:    http.StatusOK,
		Data:    customerResponse,
	})
}

// UpdateCustomer updates customer data based on ID
func UpdateCustomer(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		Username    string `json:"username" form:"username"`
		Email       string `json:"email" form:"email"`
		OldPassword string `json:"old_password" form:"old_password"`
		Password    string `json:"password" form:"password"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Format masukan tidak valid"})
		return
	}

	var customer models.User
	if err := config.DB.Where("role = ? AND id = ?", "customer", id).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Pelanggan tidak ditemukan"})
		return
	}

	if body.Username != "" {
		customer.Username = body.Username
	}
	if body.Email != "" {
		customer.Email = body.Email
	}
	if body.Password != "" {
		if body.OldPassword != "" {
			if err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(body.OldPassword)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Kata sandi lama tidak sesuai"})
				return
			}
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal memproses kata sandi"})
			return
		}
		customer.Password = string(hash)
	}

	if err := config.DB.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal memperbarui data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Data berhasil diperbarui"})
}

// DeleteCustomer deletes a customer based on ID
func DeleteCustomer(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Where("role = ?", "customer").Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

// GetCustomerLoginHistory returns login history for a given customer (admin or owning customer)
func GetCustomerLoginHistory(c *gin.Context) {
	id := c.Param("id")

	// Authorization: allow admin or the owning customer
	role, _ := c.Get("role")
	loggedInUserID, _ := c.Get("user_id")

	if roleStr, ok := role.(string); ok && roleStr == "customer" {
		if uid, ok := loggedInUserID.(uint); ok {
			if strconv.FormatUint(uint64(uid), 10) != id {
				c.JSON(http.StatusForbidden, gin.H{"message": "You don't have permission to access this resource"})
				return
			}
		}
	}

	// Pagination (optional)
	pageStr := c.Query("page")
	perPageStr := c.Query("per_page")
	page := 1
	perPage := 20
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
		perPage = pp
	}

	var histories []models.LoginHistory
	offset := (page - 1) * perPage
	if err := config.DB.Where("user_id = ?", id).Order("created_at desc").Limit(perPage).Offset(offset).Find(&histories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve login history"})
		return
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Successfully retrieved login history",
		Data:    histories,
	})
}

