package courier_controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
	"golang.org/x/crypto/bcrypt"
)

// GetCouriers retrieves all couriers
func GetCouriers(c *gin.Context) {
	role, existsRole := c.Get("role")
	loggedInUserID, existsUser := c.Get("user_id")

	query := config.DB.Where("role = ?", "courier")

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

	var couriers []models.User
	if err := query.Find(&couriers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve couriers"})
		return
	}

	var courierResponses []response.UserResponse
	for _, courier := range couriers {
		courierResponse := response.UserResponse{
			ID:       courier.ID,
			Username: courier.Username,
			Email:    courier.Email,
		}
		courierResponses = append(courierResponses, courierResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully retrieved couriers",
		"code":    http.StatusOK,
		"data":    courierResponses,
	})
}

// GetCourier retrieves a single courier based on ID
func GetCourier(c *gin.Context) {
	id := c.Param("id")

	var courier models.User
	if err := config.DB.Where("role = ? AND id = ?", "courier", id).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Courier not found"})
		return
	}

	courierResponse := response.UserResponse{
		ID:       courier.ID,
		Username: courier.Username,
		Email:    courier.Email,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully retrieved courier profile",
		"code":    http.StatusOK,
		"data":    courierResponse,
	})
}

// UpdateCourier updates a courier's profile based on ID
func UpdateCourier(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		Username string `json:"username" form:"username"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input format"})
		return
	}

	var courier models.User
	if err := config.DB.Where("role = ? AND id = ?", "courier", id).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Courier not found"})
		return
	}

	courier.Username = body.Username
	courier.Email = body.Email
	if body.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error hashing password"})
			return
		}
		courier.Password = string(hash)
	}

	if err := config.DB.Save(&courier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update courier"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Courier updated successfully"})
}

// DeleteCourier deletes a courier based on ID
func DeleteCourier(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Where("role = ? AND id = ?", "courier", id).Delete(&models.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete courier"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Courier deleted successfully"})
}
