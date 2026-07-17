package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
	"golang.org/x/crypto/bcrypt"
)

// GetUsers retrieves all users with their addresses
func GetUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Preload("Addresses").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve users"})
		return
	}

	var userResponses []response.UserResponse
	for _, user := range users {
		var addressResponses []response.AddressResponse
		for _, address := range user.Addresses {
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

		userResponse := response.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      &user.Role,
			Addresses: &addressResponses,
		}
		userResponses = append(userResponses, userResponse)
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Successfully retrieved users",
		Code:    http.StatusOK,
		Data:    userResponses,
	})
}

// GetUser retrieves a single user with their addresses based on ID
func GetUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := config.DB.Preload("Addresses").First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	var addressResponses []response.AddressResponse
	for _, address := range user.Addresses {
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

	userResponse := response.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      &user.Role,
		Addresses: &addressResponses,
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Successfully retrieved profile",
		Code:    http.StatusOK,
		Data:    userResponse,
	})
}
func UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		Username string `json:"username" form:"username"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
		Role     string `json:"role" form:"role"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input format"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// Validate role
	if body.Role != "" && !isValidRole(body.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role. Allowed roles are: admin, courier, customer"})
		return
	}

	user.Username = body.Username
	user.Email = body.Email
	if body.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error hashing password"})
			return
		}
		user.Password = string(hash)
	}
	if body.Role != "" {
		user.Role = body.Role
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// CreateUser allows an admin to create a user and explicitly set their role
func CreateUser(c *gin.Context) {
	var body struct {
		Username string `json:"username" form:"username"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
		Role     string `json:"role" form:"role"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input format"})
		return
	}

	if body.Username == "" || body.Email == "" || body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Username, email, and password are required"})
		return
	}

	// Validate role
	if body.Role == "" {
		body.Role = "customer" // Default to customer if not specified
	} else if !isValidRole(body.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role. Allowed roles are: admin, courier, customer"})
		return
	}

	// Check password length
	if len(body.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Password must be at least 8 characters long"})
		return
	}

	// Hashing the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error hashing password"})
		return
	}

	// Create user object
	user := models.User{
		Username: body.Username,
		Email:    body.Email,
		Password: string(hash),
		Role:     body.Role,
	}

	// Save user to database
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User created successfully",
		"data": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func isValidRole(role string) bool {
	return role == "admin" || role == "courier" || role == "customer"
}
