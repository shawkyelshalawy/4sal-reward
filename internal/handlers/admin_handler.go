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

func (h *AdminHandler) UpdateCreditPackage(c *gin.Context) {
	packageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package ID"})
		return
	}

	var req struct {
		Name         *string  `json:"name"`
		Description  *string  `json:"description"`
		Price        *float64 `json:"price"`
		RewardPoints *int     `json:"reward_points"`
		IsActive     *bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Price != nil && *req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than 0"})
		return
	}
	if req.RewardPoints != nil && *req.RewardPoints <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reward points must be greater than 0"})
		return
	}

	err = h.CreditService.UpdatePackage(c.Request.Context(), packageID, req.Name, req.Description, req.Price, req.RewardPoints, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Credit package updated successfully"})
}

func (h *AdminHandler) UpdateProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req struct {
		Name            *string `json:"name"`
		Description     *string `json:"description"`
		CategoryID      *string `json:"category_id"`
		PointCost       *int    `json:"point_cost"`
		StockQuantity   *int    `json:"stock_quantity"`
		IsActive        *bool   `json:"is_active"`
		IsInOfferPool   *bool   `json:"is_in_offer_pool"`
		ImageURL        *string `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.PointCost != nil && *req.PointCost <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Point cost must be greater than 0"})
		return
	}
	if req.StockQuantity != nil && *req.StockQuantity < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stock quantity cannot be negative"})
		return
	}

	var categoryID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}
		categoryID = &id
	}

	err = h.ProductService.UpdateProduct(c.Request.Context(), productID, req.Name, req.Description, categoryID, req.PointCost, req.StockQuantity, req.IsActive, req.IsInOfferPool, req.ImageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

