package controllers

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

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

	// Determine role: use body.Role if provided, default to customer
	role := body.Role
	if role == "" {
		role = "customer"
	}

	// Create user object
	user := models.User{
		Username: body.Username,
		Email:    body.Email,
		Password: string(hash),
		Role:     role,
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

func generateOTP() string {
	nBig, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "123456" // fallback
	}
	return fmt.Sprintf("%06d", nBig.Int64()+100000)
}

func ForgotPassword(c *gin.Context) {
	var body struct {
		Email string `json:"email" form:"email"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input format"})
		return
	}

	// Verify if user exists
	var user models.User
	if err := config.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Email address not registered"})
		return
	}

	// Generate 6-digit OTP
	otpCode := generateOTP()
	expiresAt := time.Now().Add(5 * time.Minute)

	// Save or overwrite OTP in database
	var existingOTP models.PasswordResetOTP
	err := config.DB.Where("email = ?", body.Email).First(&existingOTP).Error
	if err == nil {
		// Update existing record
		existingOTP.OTP = otpCode
		existingOTP.ExpiresAt = expiresAt
		config.DB.Save(&existingOTP)
	} else {
		// Create new record
		newOTP := models.PasswordResetOTP{
			Email:     body.Email,
			OTP:       otpCode,
			ExpiresAt: expiresAt,
		}
		config.DB.Create(&newOTP)
	}

	// Format HTML Email Template
	htmlMessage := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; background-color: #f4f7fc; color: #1e293b; margin: 0; padding: 20px; }
        .card { max-width: 480px; margin: 20px auto; background: #ffffff; border-radius: 16px; padding: 32px; box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05); border: 1px solid #e2e8f0; }
        .logo { text-align: center; margin-bottom: 24px; }
        .logo img { width: 140px; }
        h2 { font-size: 22px; font-weight: 700; color: #0f172a; text-align: center; margin-top: 0; margin-bottom: 8px; }
        p { font-size: 15px; line-height: 1.6; color: #475569; text-align: center; margin: 10px 0; }
        .otp-code { font-size: 34px; font-weight: 800; letter-spacing: 6px; color: #1e3a8a; text-align: center; margin: 24px 0; padding: 14px; background: #eff6ff; border-radius: 12px; border: 1px dashed #bfdbfe; font-family: monospace; }
        .footer { font-size: 11px; color: #94a3b8; text-align: center; margin-top: 32px; border-top: 1px solid #e2e8f0; padding-top: 16px; line-height: 1.4; }
    </style>
</head>
<body>
    <div class="card">
        <div class="logo">
            <img src="https://raw.githubusercontent.com/bntngridp/admin-myLaundry/main/assets/img/logo-mylaundry.png" alt="myLaundry Logo">
        </div>
        <h2>Reset Kata Sandi Anda</h2>
        <p>Halo,</p>
        <p>Kami menerima permintaan untuk mereset kata sandi akun myLaundry Anda. Gunakan kode verifikasi OTP berikut untuk melanjutkan proses pemulihan:</p>
        <div class="otp-code">%s</div>
        <p style="font-size: 13px; color: #ef4444; font-weight: 600; text-align: center;">Kode ini hanya berlaku selama 5 menit. Jangan bagikan kode ini kepada siapapun.</p>
        <div class="footer">
            Email ini dikirim secara otomatis oleh sistem keamanan myLaundry.<br>
            &copy; 2026 myLaundry. Hak Cipta Dilindungi Undang-Undang.
        </div>
    </div>
</body>
</html>
`, otpCode)

	// Send Real Email using Gmail SMTP
	err = utils.SendEmail(body.Email, "myLaundry — Kode OTP Verifikasi", htmlMessage)
	if err != nil {
		log.Println("SMTP Email Sending Failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "OTP verification code sent to your email!",
	})
}

func ResetPassword(c *gin.Context) {
	var body struct {
		Email           string `json:"email" form:"email"`
		OTP             string `json:"otp" form:"otp"`
		Password        string `json:"password" form:"password"`
		ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input format"})
		return
	}

	// Check password match
	if body.Password != body.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Passwords do not match"})
		return
	}

	// Check password strength
	if ok, errMsg := utils.ValidatePasswordStrength(body.Password); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": errMsg})
		return
	}

	// Verify OTP
	var otpRecord models.PasswordResetOTP
	if err := config.DB.Where("email = ? AND otp = ?", body.Email, body.OTP).First(&otpRecord).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid OTP code"})
		return
	}

	// Check expiration
	if time.Now().After(otpRecord.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "OTP has expired"})
		return
	}

	// Update user password
	var user models.User
	if err := config.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error hashing password"})
		return
	}

	user.Password = string(hash)
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update password"})
		return
	}

	// Delete verified OTP record
	config.DB.Unscoped().Delete(&otpRecord)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password updated successfully!",
	})
}
