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

type CreditHandler struct {
	CreditService *services.CreditService
	Logger        *logger.ZapLogger
}

type StandardResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func NewCreditHandler(creditService *services.CreditService) *CreditHandler {
	return &CreditHandler{
		CreditService: creditService,
		Logger:        logger.NewZapLogger(),
	}
}

func (h *CreditHandler) PurchaseCreditPackage(c *gin.Context) {
	requestID := generateRequestID()
	
	h.Logger.Info("Credit package purchase request started",
		zap.String("request_id", requestID),
		zap.String("client_ip", c.ClientIP()),
	)

	var req struct {
		UserID     string  `json:"user_id" binding:"required"`
		PackageID  string  `json:"package_id" binding:"required"`
		AmountPaid float64 `json:"amount_paid" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Error("Invalid request body for credit purchase",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "INVALID_REQUEST", "Invalid request format", err.Error())
		return
	}
	
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.Logger.Error("Invalid user ID format",
			zap.String("request_id", requestID),
			zap.String("user_id", req.UserID),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "INVALID_USER_ID", "Invalid user ID format", err.Error())
		return
	}
	
	packageID, err := uuid.Parse(req.PackageID)
	if err != nil {
		h.Logger.Error("Invalid package ID format",
			zap.String("request_id", requestID),
			zap.String("package_id", req.PackageID),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "INVALID_PACKAGE_ID", "Invalid package ID format", err.Error())
		return
	}
	
	if req.AmountPaid <= 0 {
		h.Logger.Error("Invalid amount paid",
			zap.String("request_id", requestID),
			zap.Float64("amount_paid", req.AmountPaid),
		)
		h.sendErrorResponse(c, requestID, "INVALID_AMOUNT", "Amount paid must be greater than 0", "")
		return
	}
	
	h.Logger.Info("Processing credit package purchase",
		zap.String("request_id", requestID),
		zap.String("user_id", req.UserID),
		zap.String("package_id", req.PackageID),
		zap.Float64("amount_paid", req.AmountPaid),
	)
	
	if err := h.CreditService.PurchasePackage(c.Request.Context(), userID, packageID, req.AmountPaid); err != nil {
		h.Logger.Error("Failed to purchase credit package",
			zap.String("request_id", requestID),
			zap.String("user_id", req.UserID),
			zap.String("package_id", req.PackageID),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "PURCHASE_FAILED", "Failed to purchase credit package", err.Error())
		return
	}
	
	h.Logger.Info("Credit package purchased successfully",
		zap.String("request_id", requestID),
		zap.String("user_id", req.UserID),
		zap.String("package_id", req.PackageID),
	)
	
	response := StandardResponse{
		Success:   true,
		Message:   "Credit package purchased successfully",
		Timestamp: time.Now(),
		RequestID: requestID,
		Data: map[string]interface{}{
			"user_id":     req.UserID,
			"package_id":  req.PackageID,
			"amount_paid": req.AmountPaid,
		},
	}
	
	c.JSON(http.StatusCreated, response)
}

func (h *CreditHandler) GetCreditPackage(c *gin.Context) {
	requestID := generateRequestID()
	
	packageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.Logger.Error("Invalid package ID in URL",
			zap.String("request_id", requestID),
			zap.String("package_id", c.Param("id")),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "INVALID_PACKAGE_ID", "Invalid package ID format", err.Error())
		return
	}
	
	pkg, err := h.CreditService.GetPackage(c.Request.Context(), packageID)
	if err != nil {
		h.Logger.Error("Failed to fetch credit package",
			zap.String("request_id", requestID),
			zap.String("package_id", packageID.String()),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "PACKAGE_NOT_FOUND", "Credit package not found", err.Error())
		return
	}
	
	response := StandardResponse{
		Success:   true,
		Message:   "Credit package retrieved successfully",
		Data:      pkg,
		Timestamp: time.Now(),
		RequestID: requestID,
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *CreditHandler) GetCreditPackages(c *gin.Context) {
	requestID := generateRequestID()
	
	page := 1
	size := 10
	
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	
	if s := c.Query("size"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed > 0 && parsed <= 100 {
			size = parsed
		}
	}

	h.Logger.Info("Fetching credit packages",
		zap.String("request_id", requestID),
		zap.Int("page", page),
		zap.Int("size", size),
	)

	packages, total, err := h.CreditService.GetPackages(c.Request.Context(), page, size)
	if err != nil {
		h.Logger.Error("Failed to fetch credit packages",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "PACKAGES_FETCH_ERROR", "Failed to fetch credit packages", err.Error())
		return
	}

	h.Logger.Info("Credit packages fetched successfully",
		zap.String("request_id", requestID),
		zap.Int("package_count", len(packages)),
		zap.Int("total", total),
	)

	response := StandardResponse{
		Success:   true,
		Message:   "Credit packages retrieved successfully",
		Timestamp: time.Now(),
		RequestID: requestID,
		Data: map[string]interface{}{
			"packages":     packages,
			"page":         page,
			"size":         size,
			"total":        total,
			"total_pages":  (total + size - 1) / size,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *CreditHandler) sendErrorResponse(c *gin.Context, requestID, code, message, details string) {
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
	case "INVALID_REQUEST", "INVALID_USER_ID", "INVALID_PACKAGE_ID", "INVALID_AMOUNT":
		statusCode = http.StatusBadRequest
	case "PACKAGE_NOT_FOUND":
		statusCode = http.StatusNotFound
	}
	
	c.JSON(statusCode, response)
}