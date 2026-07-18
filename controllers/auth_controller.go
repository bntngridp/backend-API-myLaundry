package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/utils"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var body struct {
		Username        string `json:"username" form:"username"`
		Email           string `json:"email" form:"email"`
		Password        string `json:"password" form:"password"`
		ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
		Role            string `json:"role" form:"role"`
	}

	// Binding request body into the struct
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input format"})
		return
	}

	// Validate password strength
	if ok, errMsg := utils.ValidatePasswordStrength(body.Password); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": errMsg})
		return
	}

	// Check password confirmation
	if body.Password != body.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Passwords do not match"})
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
		Role:     "customer", // Always default public registration to customer for security
	}

	// Save user to database
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create user"})
		return
	}

	// Respond with success message
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User created successfully",
	})
}

func Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input format"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid email or password"})
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role) // Tambahkan role ke token
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful!",
		"code":    http.StatusOK,
		"data": gin.H{
			"token": token,
			"role":  user.Role,
		},
	})
}
