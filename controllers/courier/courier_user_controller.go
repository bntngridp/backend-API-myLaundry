package courier_controllers

import (
	"net/http"
	"strconv"

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
		Username    string `json:"username" form:"username"`
		Email       string `json:"email" form:"email"`
		OldPassword string `json:"old_password" form:"old_password"`
		Password    string `json:"password" form:"password"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Format masukan tidak valid"})
		return
	}

	var courier models.User
	if err := config.DB.Where("role = ? AND id = ?", "courier", id).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Kurir tidak ditemukan"})
		return
	}

	if body.Username != "" {
		courier.Username = body.Username
	}
	if body.Email != "" {
		courier.Email = body.Email
	}
	if body.Password != "" {
		if body.OldPassword != "" {
			if err := bcrypt.CompareHashAndPassword([]byte(courier.Password), []byte(body.OldPassword)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Kata sandi lama tidak sesuai"})
				return
			}
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal memproses kata sandi"})
			return
		}
		courier.Password = string(hash)
	}

	if err := config.DB.Save(&courier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal memperbarui data kurir"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Data kurir berhasil diperbarui"})
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

// GetCourierLoginHistory returns login history for a given courier (admin or owning courier)
func GetCourierLoginHistory(c *gin.Context) {
	id := c.Param("id")

	// Authorization: allow admin or the owning courier
	role, _ := c.Get("role")
	loggedInUserID, _ := c.Get("user_id")

	if roleStr, ok := role.(string); ok && roleStr == "courier" {
		if uid, ok := loggedInUserID.(uint); ok {
			if strconv.FormatUint(uint64(uid), 10) != id {
				c.JSON(http.StatusForbidden, gin.H{"message": "You don't have permission to access this resource"})
				return
			}
		}
	}

	// Pagination
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

