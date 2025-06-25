package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/services"
)

type AdminHandler struct {
	ProductService *services.ProductService
	CreditService  *services.CreditService
}

// Create credit package
func (h *AdminHandler) CreateCreditPackage(c *gin.Context) {
	var req struct {
		Name         string  `json:"name" binding:"required"`
		Description  string  `json:"description"`
		Price        float64 `json:"price" binding:"required,gt=0"`
		RewardPoints int     `json:"reward_points" binding:"required,gt=0"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	
	packageID, err := h.CreditService.CreatePackage(c.Request.Context(), req.Name, req.Description, req.Price, req.RewardPoints)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"package_id": packageID})
}

// Create product
func (h *AdminHandler) CreateProduct(c *gin.Context) {
	var req struct {
		Name            string `json:"name" binding:"required"`
		Description     string `json:"description"`
		CategoryID      string `json:"category_id"`
		PointCost       int    `json:"point_cost" binding:"required,gt=0"`
		StockQuantity   int    `json:"stock_quantity" binding:"required,gte=0"`
		IsInOfferPool   bool   `json:"is_in_offer_pool"`
		ImageURL        string `json:"image_url"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	
	var categoryID *uuid.UUID
	if req.CategoryID != "" {
		id, err := uuid.Parse(req.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}
		categoryID = &id
	}
	
	productID, err := h.ProductService.CreateProduct(c.Request.Context(), req.Name, req.Description, categoryID, req.PointCost, req.StockQuantity, req.IsInOfferPool, req.ImageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"product_id": productID})
}

// Update product offer pool status
func (h *AdminHandler) UpdateProductOfferStatus(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	
	var req struct {
		IsInOfferPool bool `json:"is_in_offer_pool"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	
	if err := h.ProductService.UpdateOfferStatus(c.Request.Context(), productID, req.IsInOfferPool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Status(http.StatusOK)
}