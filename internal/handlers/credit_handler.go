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
		UserID     string  `json:"user_id" binding:"required"`
		PackageID  string  `json:"package_id" binding:"required"`
		AmountPaid float64 `json:"amount_paid" binding:"required"`
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
	
	if req.AmountPaid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount paid must be greater than 0"})
		return
	}
	
	if err := h.CreditService.PurchasePackage(c.Request.Context(), userID, packageID, req.AmountPaid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"message": "Credit package purchased successfully"})
}

func (h *CreditHandler) GetCreditPackage(c *gin.Context) {
	packageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package ID"})
		return
	}
	
	pkg, err := h.CreditService.GetPackage(c.Request.Context(), packageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, pkg)
}