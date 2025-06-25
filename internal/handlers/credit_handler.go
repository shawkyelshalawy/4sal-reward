package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/services"
)




type CreditHandler struct {
	CreditService *services.CreditService
}

func (h *CreditHandler) PurchaseCreditPackage(c *gin.Context) {
	var req struct {
		UserID    string `json:"user_id" binding:"required"`
		PackageID string `json:"package_id" binding:"required"`
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
	
	packageID, err := uuid.Parse(req.PackageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package ID"})
		return
	}
	
	if err := h.CreditService.PurchasePackage(c.Request.Context(), userID, packageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Status(http.StatusCreated)
}