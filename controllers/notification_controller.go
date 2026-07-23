package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
)

func GetNotifications(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		// Float64 fallback if parsed from JWT claims
		if floatVal, floatOk := userIDVal.(float64); floatOk {
			userID = uint(floatVal)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user session"})
			return
		}
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	var notifications []models.Notification
	config.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&notifications)

	// If user has no notifications yet, generate role-specific default notifications
	if len(notifications) == 0 {
		var initialNotifs []models.Notification

		if user.Role == "courier" {
			initialNotifs = []models.Notification{
				{
					UserID:  userID,
					Title:   "Selamat Datang di myLaundry Kurir! 🚚",
					Message: "Siapkan kendaraanmu dan pastikan status siap untuk menerima pesanan penjemputan.",
					Type:    "info",
					IsRead:  false,
				},
				{
					UserID:  userID,
					Title:   "Tips Pelayanan Kurir ✨",
					Message: "Selalu konfirmasi jumlah dan kondisi pakaian saat melakukan penjemputan dari pelanggan.",
					Type:    "info",
					IsRead:  false,
				},
			}

			// Check for pending orders to assign as tasks
			var pendingOrders []models.Order
			config.DB.Where("status IN ?", []string{"pending", "accepted", "delivering"}).Limit(3).Find(&pendingOrders)
			for _, ord := range pendingOrders {
				if ord.Status == "pending" {
					initialNotifs = append(initialNotifs, models.Notification{
						UserID:  userID,
						Title:   fmt.Sprintf("Pesanan Penjemputan #%d 📍", ord.ID),
						Message: "Ada pesanan pelanggan baru yang membutuhkan penjemputan. Segera terima tugas!",
						Type:    "task",
						IsRead:  false,
					})
				} else if ord.CourierID != nil && *ord.CourierID == userID {
					initialNotifs = append(initialNotifs, models.Notification{
						UserID:  userID,
						Title:   fmt.Sprintf("Tugas Pesanan #%d Terbuka", ord.ID),
						Message: fmt.Sprintf("Status pesanan saat ini: %s. Periksa detail pesanan di aplikasi.", ord.Status),
						Type:    "order_status",
						IsRead:  false,
					})
				}
			}
		} else { // Customer default notifications
			initialNotifs = []models.Notification{
				{
					UserID:  userID,
					Title:   "Promo Spesial #BersihTanpaPusing 🎟️",
					Message: "Gunakan kode promo BersihTanpaPusing untuk mendapatkan diskon 30% pada pemesanan pertamamu!",
					Type:    "promo",
					IsRead:  false,
				},
				{
					UserID:  userID,
					Title:   "Selamat Datang di myLaundry! ✨",
					Message: "Nikmati kemudahan layanan cuci lipat, cuci setrika, dan jemput antar langsung ke lokasi rumahmu.",
					Type:    "info",
					IsRead:  false,
				},
			}

			// Check customer's active orders
			var customerOrders []models.Order
			config.DB.Where("customer_id = ?", userID).Order("id desc").Limit(2).Find(&customerOrders)
			for _, ord := range customerOrders {
				initialNotifs = append(initialNotifs, models.Notification{
					UserID:  userID,
					Title:   fmt.Sprintf("Status Pesanan #%d 🧺", ord.ID),
					Message: fmt.Sprintf("Pesanan laundry kamu saat ini dalam status: %s.", ord.Status),
					Type:    "order_status",
					IsRead:  false,
				})
			}
		}

		for i := range initialNotifs {
			config.DB.Create(&initialNotifs[i])
		}
		notifications = initialNotifs
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notifications fetched successfully",
		"data":    notifications,
	})
}

func MarkNotificationAsRead(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	notifID := c.Param("id")
	var notif models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", notifID, userIDVal).First(&notif).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Notification not found"})
		return
	}

	notif.IsRead = true
	config.DB.Save(&notif)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notification marked as read",
	})
}

func MarkAllNotificationsAsRead(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	config.DB.Model(&models.Notification{}).Where("user_id = ?", userIDVal).Update("is_read", true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All notifications marked as read",
	})
}
