package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/repositories"
	"github.com/shawkyelshalawy/4sal-reward/internal/services"
)

type AIHandler struct {
	ProductService *services.ProductService
	UserRepo       *repositories.UserRepository
}

func (h *AIHandler) GetRecommendation(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	// Get user's point balance
	user, err := h.UserRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}
	
	recommendation := h.getSimpleRecommendation(user.PointBalance)
	
	c.JSON(http.StatusOK, gin.H{
		"user_id":         userID,
		"point_balance":   user.PointBalance,
		"recommendation": recommendation,
		"reason":         "Based on your current point balance and popular products",
	})
}

// Simple recommendation logic (replace with actual AI service call)
func (h *AIHandler) getSimpleRecommendation(pointBalance int) map[string]interface{} {
	if pointBalance >= 1000 {
		return map[string]interface{}{
			"message":    "You have enough points for premium products! Consider high-value electronics or gadgets.",
			"category":   "Electronics",
			"min_points": 500,
			"max_points": 1500,
		}
	} else if pointBalance >= 500 {
		return map[string]interface{}{
			"message":    "Great! You can redeem mid-range products like accessories or books.",
			"category":   "Accessories",
			"min_points": 200,
			"max_points": 600,
		}
	} else if pointBalance >= 100 {
		return map[string]interface{}{
			"message":    "You can redeem small items or save up for bigger rewards!",
			"category":   "Small Items",
			"min_points": 50,
			"max_points": 150,
		}
	} else {
		return map[string]interface{}{
			"message":    "Keep earning points! Consider purchasing more credit packages.",
			"category":   "Credit Packages",
			"suggestion": "Buy more credits to earn reward points",
		}
	}
}