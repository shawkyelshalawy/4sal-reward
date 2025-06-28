package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/infrastructure/logger"
	"github.com/shawkyelshalawy/4sal-reward/internal/services"
	"go.uber.org/zap"
)

type ProductHandler struct {
	ProductService *services.ProductService
	Logger         *logger.ZapLogger
}

func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{
		ProductService: productService,
		Logger:         logger.NewZapLogger(),
	}
}

func (h *ProductHandler) RedeemProduct(c *gin.Context) {
	requestID := generateRequestID()
	
	h.Logger.Info("Product redemption request started",
		zap.String("request_id", requestID),
		zap.String("client_ip", c.ClientIP()),
	)

	var req struct {
		UserID    string `json:"user_id" binding:"required"`
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,min=1"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Error("Invalid request body for product redemption",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		h.sendProductErrorResponse(c, requestID, "INVALID_REQUEST", "Invalid request format", err.Error())
		return
	}
	
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.Logger.Error("Invalid user ID format",
			zap.String("request_id", requestID),
			zap.String("user_id", req.UserID),
			zap.Error(err),
		)
		h.sendProductErrorResponse(c, requestID, "INVALID_USER_ID", "Invalid user ID format", err.Error())
		return
	}
	
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		h.Logger.Error("Invalid product ID format",
			zap.String("request_id", requestID),
			zap.String("product_id", req.ProductID),
			zap.Error(err),
		)
		h.sendProductErrorResponse(c, requestID, "INVALID_PRODUCT_ID", "Invalid product ID format", err.Error())
		return
	}
	
	h.Logger.Info("Processing product redemption",
		zap.String("request_id", requestID),
		zap.String("user_id", req.UserID),
		zap.String("product_id", req.ProductID),
		zap.Int("quantity", req.Quantity),
	)
	
	if err := h.ProductService.RedeemProduct(c.Request.Context(), userID, productID, req.Quantity); err != nil {
		h.Logger.Error("Failed to redeem product",
			zap.String("request_id", requestID),
			zap.String("user_id", req.UserID),
			zap.String("product_id", req.ProductID),
			zap.Int("quantity", req.Quantity),
			zap.Error(err),
		)
		h.sendProductErrorResponse(c, requestID, "REDEMPTION_FAILED", "Failed to redeem product", err.Error())
		return
	}
	
	h.Logger.Info("Product redeemed successfully",
		zap.String("request_id", requestID),
		zap.String("user_id", req.UserID),
		zap.String("product_id", req.ProductID),
		zap.Int("quantity", req.Quantity),
	)
	
	response := StandardResponse{
		Success:   true,
		Message:   "Product redeemed successfully",
		Timestamp: time.Now(),
		RequestID: requestID,
		Data: map[string]interface{}{
			"user_id":    req.UserID,
			"product_id": req.ProductID,
			"quantity":   req.Quantity,
		},
	}
	
	c.JSON(http.StatusCreated, response)
}

func (h *ProductHandler) SearchProducts(c *gin.Context) {
	requestID := generateRequestID()
	
	query := c.Query("query")
	if query == "" {
		h.Logger.Error("Missing query parameter",
			zap.String("request_id", requestID),
		)
		h.sendProductErrorResponse(c, requestID, "MISSING_QUERY", "Query parameter is required", "")
		return
	}
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	
	h.Logger.Info("Processing product search",
		zap.String("request_id", requestID),
		zap.String("query", query),
		zap.Int("page", page),
		zap.Int("size", size),
	)
	
	products, err := h.ProductService.SearchProducts(c.Request.Context(), query, page, size)
	if err != nil {
		h.Logger.Error("Failed to search products",
			zap.String("request_id", requestID),
			zap.String("query", query),
			zap.Error(err),
		)
		h.sendProductErrorResponse(c, requestID, "SEARCH_FAILED", "Failed to search products", err.Error())
		return
	}
	
	h.Logger.Info("Product search completed",
		zap.String("request_id", requestID),
		zap.String("query", query),
		zap.Int("results_count", len(products)),
	)
	
	response := StandardResponse{
		Success:   true,
		Message:   "Product search completed successfully",
		Timestamp: time.Now(),
		RequestID: requestID,
		Data: map[string]interface{}{
			"products": products,
			"query":    query,
			"page":     page,
			"size":     size,
		},
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *ProductHandler) sendProductErrorResponse(c *gin.Context, requestID, code, message, details string) {
	response := StandardResponse{
		Success:   false,
		Message:   "Request failed",
		Timestamp: time.Now(),
		RequestID: requestID,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	
	statusCode := http.StatusInternalServerError
	switch code {
	case "INVALID_REQUEST", "INVALID_USER_ID", "INVALID_PRODUCT_ID", "MISSING_QUERY":
		statusCode = http.StatusBadRequest
	case "PRODUCT_NOT_FOUND":
		statusCode = http.StatusNotFound
	}
	
	c.JSON(statusCode, response)
}