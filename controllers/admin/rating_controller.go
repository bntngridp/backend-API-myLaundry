package admin_controllers

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raihansyahrin/backend_laundry_app.git/config"
	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"github.com/raihansyahrin/backend_laundry_app.git/response"
)

type ReplyRatingRequest struct {
	AdminReply string `json:"admin_reply" binding:"required"`
}

// GetAllRatings - Get all customer ratings & reviews with filters & stats
func GetAllRatings(c *gin.Context) {
	scoreFilter := c.Query("score")
	searchQuery := c.Query("search")

	db := config.DB.Model(&models.Rating{}).Preload("Customer").Preload("Courier").Preload("Branch")

	if scoreFilter != "" {
		scoreNum, err := strconv.ParseFloat(scoreFilter, 64)
		if err == nil {
			db = db.Where("FLOOR(branch_score) = ?", scoreNum)
		}
	}

	if searchQuery != "" {
		likeQuery := "%" + searchQuery + "%"
		db = db.Where("review_text LIKE ? OR tags LIKE ?", likeQuery, likeQuery)
	}

	var ratings []models.Rating
	if err := db.Order("id desc").Find(&ratings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Code:    http.StatusInternalServerError,
			Success: false,
			Message: "Gagal mengambil daftar ulasan: " + err.Error(),
		})
		return
	}

	// Calculate overall statistics
	var totalRatings int64
	var avgBranchScore float64
	var avgCourierScore float64
	var score5Count, score4Count, score3Count, score2Count, score1Count int64

	config.DB.Model(&models.Rating{}).Count(&totalRatings)
	config.DB.Model(&models.Rating{}).Select("COALESCE(AVG(branch_score), 5.0)").Scan(&avgBranchScore)
	config.DB.Model(&models.Rating{}).Select("COALESCE(AVG(courier_score), 5.0)").Scan(&avgCourierScore)

	config.DB.Model(&models.Rating{}).Where("FLOOR(branch_score) = 5").Count(&score5Count)
	config.DB.Model(&models.Rating{}).Where("FLOOR(branch_score) = 4").Count(&score4Count)
	config.DB.Model(&models.Rating{}).Where("FLOOR(branch_score) = 3").Count(&score3Count)
	config.DB.Model(&models.Rating{}).Where("FLOOR(branch_score) = 2").Count(&score2Count)
	config.DB.Model(&models.Rating{}).Where("FLOOR(branch_score) = 1").Count(&score1Count)

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Daftar ulasan berhasil diambil",
		Data: map[string]interface{}{
			"ratings": ratings,
			"stats": map[string]interface{}{
				"total_ratings":     totalRatings,
				"avg_branch_score":  math.Round(avgBranchScore*10) / 10,
				"avg_courier_score": math.Round(avgCourierScore*10) / 10,
				"score_5_count":     score5Count,
				"score_4_count":     score4Count,
				"score_3_count":     score3Count,
				"score_2_count":     score2Count,
				"score_1_count":     score1Count,
			},
		},
	})
}

// ReplyRating - Submit official Admin response to a customer review
func ReplyRating(c *gin.Context) {
	ratingIDParam := c.Param("id")
	ratingID, err := strconv.Atoi(ratingIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "ID Rating tidak valid",
		})
		return
	}

	var req ReplyRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultResponse{
			Code:    http.StatusBadRequest,
			Success: false,
			Message: "Payload balasan ulasan tidak valid: " + err.Error(),
		})
		return
	}

	var rating models.Rating
	if err := config.DB.First(&rating, ratingID).Error; err != nil {
		c.JSON(http.StatusNotFound, response.DefaultResponse{
			Code:    http.StatusNotFound,
			Success: false,
			Message: "Ulasan tidak ditemukan",
		})
		return
	}

	now := time.Now()
	rating.AdminReply = req.AdminReply
	rating.RepliedAt = &now

	if err := config.DB.Save(&rating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultResponse{
			Code:    http.StatusInternalServerError,
			Success: false,
			Message: "Gagal menyimpan balasan ulasan",
		})
		return
	}

	// Preload relationships for response
	config.DB.Preload("Customer").Preload("Courier").Preload("Branch").First(&rating, rating.ID)

	c.JSON(http.StatusOK, response.DefaultResponse{
		Code:    http.StatusOK,
		Success: true,
		Message: "Balasan ulasan berhasil disimpan",
		Data:    rating,
	})
}
