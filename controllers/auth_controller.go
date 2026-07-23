package controllers

import (
	"crypto/rand"
	"encoding/json"
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
		PhoneNumber     string `json:"phone_number" form:"phone_number"`
		Password        string `json:"password" form:"password"`
		ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
		Role            string `json:"role" form:"role"`
		EmployeeCode    string `json:"employee_code" form:"employee_code"`
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
		Username:    body.Username,
		Email:       body.Email,
		PhoneNumber: body.PhoneNumber,
		Password:    string(hash),
		Role:        role,
	}

	// Handle courier specific registration details
	if role == "courier" {
		var adminID uint
		if body.EmployeeCode == "EBS-f4wD" || body.EmployeeCode == "2" || body.EmployeeCode == "admin2" || body.EmployeeCode == "EBS-admin2" {
			adminID = 2
		} else if body.EmployeeCode == "EBS-admin1" || body.EmployeeCode == "1" || body.EmployeeCode == "admin1" {
			adminID = 1
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Kode Khusus Karyawan tidak valid"})
			return
		}
		user.CreatedByAdminID = &adminID
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
	if err := config.DB.Where("email = ? OR phone_number = ?", body.Email, body.Email).First(&user).Error; err != nil {
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
	nBig, err := rand.Int(rand.Reader, big.NewInt(9000))
	if err != nil {
		return "1234" // fallback
	}
	return fmt.Sprintf("%04d", nBig.Int64()+1000)
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
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Kode OTP Pemulihan Kata Sandi - myLaundry</title>
    <style>
        body {
            font-family: 'Outfit', 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background-color: #f8fafc;
            color: #0f172a;
            margin: 0;
            padding: 0;
            -webkit-font-smoothing: antialiased;
        }
        .wrapper {
            width: 100%%;
            background-color: #f8fafc;
            padding: 40px 20px;
        }
        .container {
            max-width: 520px;
            margin: 0 auto;
            background-color: #ffffff;
            border-radius: 20px;
            overflow: hidden;
            box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.05), 0 8px 10px -6px rgba(0, 0, 0, 0.05);
            border: 1px solid #e2e8f0;
        }
        .header {
            background: linear-gradient(135deg, #0B1739 0%%, #1e293b 100%%);
            padding: 40px 32px;
            text-align: center;
        }
        .logo {
            height: 48px;
            margin-bottom: 8px;
        }
        .content {
            padding: 40px 36px;
            text-align: center;
        }
        h1 {
            font-size: 24px;
            font-weight: 700;
            color: #0f172a;
            margin-top: 0;
            margin-bottom: 16px;
        }
        p {
            font-size: 15px;
            line-height: 1.6;
            color: #475569;
            margin: 0 0 24px 0;
        }
        .otp-box {
            background-color: #f1f5f9;
            border: 2px dashed #cbd5e1;
            border-radius: 16px;
            padding: 20px;
            margin: 32px 0;
            text-align: center;
        }
        .otp-code {
            font-size: 38px;
            font-weight: 800;
            letter-spacing: 8px;
            color: #0d6efd;
            font-family: 'Courier New', Courier, monospace;
            margin: 0;
        }
        .warning-text {
            font-size: 13px;
            color: #ef4444;
            font-weight: 600;
            margin-bottom: 0;
            background-color: #fef2f2;
            border-radius: 8px;
            padding: 8px 12px;
            display: inline-block;
        }
        .footer {
            background-color: #f8fafc;
            padding: 24px 32px;
            text-align: center;
            border-top: 1px solid #f1f5f9;
        }
        .footer-text {
            font-size: 12px;
            color: #94a3b8;
            line-height: 1.5;
            margin: 0;
        }
        .footer-links {
            margin-top: 12px;
        }
        .footer-links a {
            color: #64748b;
            text-decoration: none;
            font-size: 12px;
            margin: 0 8px;
        }
        .footer-links a:hover {
            color: #0d6efd;
        }
    </style>
</head>
<body>
    <div class="wrapper">
        <div class="container">
            <!-- Header section with dark background & logo -->
            <div class="header">
                <img class="logo" src="https://raw.githubusercontent.com/bntngridp/admin-myLaundry/main/assets/img/logo-nobg.png" alt="myLaundry Logo">
            </div>
            
            <!-- Content section -->
            <div class="content">
                <h1>Reset Kata Sandi</h1>
                <p>Halo,</p>
                <p>Kami menerima permintaan untuk mengatur ulang kata sandi akun <strong>myLaundry Admin</strong> Anda. Silakan gunakan kode verifikasi OTP di bawah ini untuk melanjutkan:</p>
                
                <div class="otp-box">
                    <div class="otp-code">%s</div>
                </div>
                
                <p class="warning-text">⚠️ Kode OTP ini berlaku selama 5 menit. Jangan bagikan kode ini kepada siapapun.</p>
            </div>
            
            <!-- Footer section -->
            <div class="footer">
                <p class="footer-text">Email ini dikirim secara otomatis oleh sistem keamanan myLaundry. Jika Anda tidak merasa melakukan permintaan ini, silakan abaikan email ini.</p>
                <div class="footer-links">
                    <a href="#">Bantuan</a>
                    &middot;
                    <a href="#">Kebijakan Privasi</a>
                </div>
                <p class="footer-text" style="margin-top: 16px;">&copy; 2026 myLaundry Admin. Hak Cipta Dilindungi.</p>
            </div>
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

func VerifyOTP(c *gin.Context) {
	var body struct {
		Email string `json:"email" form:"email"`
		OTP   string `json:"otp" form:"otp"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input format"})
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "OTP verified successfully!",
	})
}

func GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User profile retrieved successfully",
		"data": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func GoogleLogin(c *gin.Context) {
	var body struct {
		IDToken string `json:"id_token" form:"id_token"`
		Role    string `json:"role" form:"role"`
	}

	if err := c.ShouldBind(&body); err != nil || body.IDToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID token required"})
		return
	}

	// Verify ID Token against Google TokenInfo endpoint
	resp, err := http.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", body.IDToken))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid Google token"})
		return
	}
	defer resp.Body.Close()

	var googleClaims struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		EmailVerified string `json:"email_verified"`
		Sub           string `json:"sub"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleClaims); err != nil || googleClaims.Email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Failed to decode Google profile"})
		return
	}

	role := body.Role
	if role == "" {
		role = "customer"
	}

	// Find or create user
	var user models.User
	if err := config.DB.Where("email = ?", googleClaims.Email).First(&user).Error; err != nil {
		// User does not exist, create new user
		username := googleClaims.Name
		if username == "" {
			username = googleClaims.Email
		}
		dummyPass, _ := bcrypt.GenerateFromPassword([]byte("google_oauth_"+googleClaims.Sub), bcrypt.DefaultCost)
		user = models.User{
			Username: username,
			Email:    googleClaims.Email,
			Password: string(dummyPass),
			Role:     role,
		}
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create user account"})
			return
		}
	}

	// Generate JWT Token
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate session token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Google login successful!",
		"code":    http.StatusOK,
		"data": gin.H{
			"token": token,
			"role":  user.Role,
		},
	})
}
