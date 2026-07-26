package controllers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
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

	// Security: only an authenticated admin can create other admin accounts
	if role == "admin" {
		// /auth/register does not use AuthMiddleware, so we must manually validate the token
		authHeader := c.GetHeader("Authorization")
		isAdmin := false
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if tokenString != "" {
				if claims, err := utils.ValidateJWT(tokenString); err == nil && claims.Role == "admin" {
					isAdmin = true
				}
			}
		}
		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"message": "Only admin can register a new admin account"})
			return
		}
	}

	phoneNumber := strings.TrimSpace(body.PhoneNumber)
	if phoneNumber == "" {
		phoneNumber = fmt.Sprintf("08%d", time.Now().UnixNano()%10000000000)
	}

	// Check if email or phone number is already registered
	cleanEmail := strings.TrimSpace(body.Email)
	cleanPhone := strings.TrimSpace(body.PhoneNumber)

	var emailExists bool
	var phoneExists bool

	var tempUser models.User
	if cleanEmail != "" && config.DB.Where("email = ?", cleanEmail).First(&tempUser).Error == nil {
		emailExists = true
	}

	if cleanPhone != "" && config.DB.Where("phone_number = ?", cleanPhone).First(&tempUser).Error == nil {
		phoneExists = true
	}

	if emailExists && phoneExists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email dan nomor telepon sudah terdaftar, silakan gunakan email & nomor lain atau login"})
		return
	}
	if emailExists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email sudah terdaftar, silakan gunakan email lain atau login"})
		return
	}
	if phoneExists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Nomor telepon sudah terdaftar, silakan gunakan nomor lain"})
		return
	}

	// Create user object
	user := models.User{
		Username:    body.Username,
		Email:       cleanEmail,
		PhoneNumber: phoneNumber,
		Password:    string(hash),
		Role:        role,
	}

	// Handle courier specific registration details
	if role == "courier" {
		var adminID uint = 1
		if body.EmployeeCode == "EBS-f4wD" || body.EmployeeCode == "2" || body.EmployeeCode == "admin2" || body.EmployeeCode == "EBS-admin2" {
			adminID = 2
		} else if body.EmployeeCode == "EBS-admin1" || body.EmployeeCode == "1" || body.EmployeeCode == "admin1" {
			adminID = 1
		} else {
			if reqUserID, exists := c.Get("user_id"); exists {
				if uid, ok := reqUserID.(uint); ok {
					adminID = uid
				}
			}
		}
		user.CreatedByAdminID = &adminID
	}

	// Save user to database
	if err := config.DB.Create(&user).Error; err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "1062") || strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique") {
			if strings.Contains(errStr, "phone_number") || strings.Contains(errStr, "phone") {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Nomor telepon sudah terdaftar, silakan gunakan nomor lain"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"message": "Email sudah terdaftar, silakan gunakan email lain atau login"})
			return
		}
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
		Role     string `json:"role" form:"role"` // Expected role e.g. "courier", "customer", "admin"
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

	// Security: Validate expected role if specified
	expectedRole := strings.TrimSpace(strings.ToLower(body.Role))
	if expectedRole != "" && strings.ToLower(user.Role) != expectedRole {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": fmt.Sprintf("Akses ditolak: Akun Anda terdaftar sebagai %s, tidak dapat masuk ke aplikasi %s.", user.Role, expectedRole),
		})
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role) // Tambahkan role ke token
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate token"})
		return
	}

	// Record login history
	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	history := models.LoginHistory{
		UserID:    user.ID,
		Role:      user.Role,
		IP:        ip,
		UserAgent: ua,
		Success:   true,
		LoggedAt:  time.Now(),
	}
	_ = config.DB.Create(&history).Error

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
		return "123456" // fallback 6-digit
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

	cleanEmail := strings.TrimSpace(body.Email)
	if cleanEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email address is required"})
		return
	}

	// Verify if user exists
	var user models.User
	if err := config.DB.Where("email = ?", cleanEmail).First(&user).Error; err != nil {
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
		// Check rate limit: 60 seconds cooldown between OTP requests
		elapsed := time.Since(existingOTP.UpdatedAt)
		if elapsed < 60*time.Second {
			remainingSeconds := int(math.Ceil(60.0 - elapsed.Seconds()))
			if remainingSeconds < 1 {
				remainingSeconds = 1
			}
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":           false,
				"message":           fmt.Sprintf("Harap tunggu %d detik sebelum meminta kode OTP kembali.", remainingSeconds),
				"remaining_seconds": remainingSeconds,
			})
			return
		}
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
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"role":         user.Role,
			"is_available": user.IsAvailable,
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

	var googleClaims struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		EmailVerified string `json:"email_verified"`
		Sub           string `json:"sub"`
	}

	// 1. Try id_token endpoint
	resp, err := http.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", body.IDToken))
	if err == nil && resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(&googleClaims)
		resp.Body.Close()
	} else if resp != nil {
		resp.Body.Close()
	}

	// 2. If email is empty, try access_token endpoint
	if googleClaims.Email == "" {
		resp, err = http.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?access_token=%s", body.IDToken))
		if err == nil && resp.StatusCode == http.StatusOK {
			json.NewDecoder(resp.Body).Decode(&googleClaims)
			resp.Body.Close()
		} else if resp != nil {
			resp.Body.Close()
		}
	}

	// 3. If email is still empty, try userinfo endpoint with Bearer token
	if googleClaims.Email == "" {
		req, errReq := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
		if errReq == nil {
			req.Header.Set("Authorization", "Bearer "+body.IDToken)
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err = client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				json.NewDecoder(resp.Body).Decode(&googleClaims)
				resp.Body.Close()
			} else if resp != nil {
				resp.Body.Close()
			}
		}
	}

	if googleClaims.Email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid Google token"})
		return
	}

	role := body.Role
	if role == "" {
		role = "customer"
	}

	// Find or create user
	var user models.User
	if err := config.DB.Where("email = ?", googleClaims.Email).First(&user).Error; err != nil {
		// User does not exist, create new user with requested role
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
	} else {
		// User exists: enforce role matching
		if role != "" && strings.ToLower(user.Role) != strings.ToLower(role) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": fmt.Sprintf("Akses ditolak: Akun Google Anda terdaftar sebagai %s, tidak dapat masuk ke aplikasi %s.", user.Role, role),
			})
			return
		}
	}

	// Generate JWT Token
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate session token"})
		return
	}

			// Record login history for Google login
			ip := c.ClientIP()
			ua := c.Request.UserAgent()
			history := models.LoginHistory{
				UserID:    user.ID,
				Role:      user.Role,
				IP:        ip,
				UserAgent: ua,
				Success:   true,
				LoggedAt:  time.Now(),
			}
			_ = config.DB.Create(&history).Error

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

// GetLoginHistory retrieves the authenticated user's login history logs
func GetLoginHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.DefaultResponse{
			Success: false,
			Message: "Unauthorized access",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	var histories []models.LoginHistory
	if err := config.DB.Where("user_id = ?", userID).Order("logged_at desc").Limit(20).Find(&histories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Success: false,
			Message: "Failed to retrieve login history",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.DefaultResponse{
		Success: true,
		Message: "Successfully retrieved login history",
		Code:    http.StatusOK,
		Data:    histories,
	})
}


