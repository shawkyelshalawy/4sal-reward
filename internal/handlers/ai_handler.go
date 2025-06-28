package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/infrastructure/logger"
	"github.com/shawkyelshalawy/4sal-reward/internal/repositories"
	"github.com/shawkyelshalawy/4sal-reward/internal/services"
	"go.uber.org/zap"
)

type AIHandler struct {
	ProductService *services.ProductService
	UserRepo       *repositories.UserRepository
	CategoryRepo   *repositories.CategoryRepository
	Logger         *logger.ZapLogger
}

type AIRecommendationRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type AIRecommendationResponse struct {
	Success               bool                   `json:"success"`
	Message               string                 `json:"message"`
	Data                  *RecommendationData    `json:"data,omitempty"`
	Error                 *ErrorDetails          `json:"error,omitempty"`
	Timestamp             time.Time              `json:"timestamp"`
	RequestID             string                 `json:"request_id"`
}

type RecommendationData struct {
	UserID                string                 `json:"user_id"`
	PointBalance          int                    `json:"point_balance"`
	RecommendedCategoryID string                 `json:"recommended_category_id"`
	CategoryName          string                 `json:"category_name"`
	MinPointsLLM          int                    `json:"min_points_llm"`
	MaxPointsLLM          int                    `json:"max_points_llm"`
	Reasoning             string                 `json:"reasoning"`
	RecommendedProducts   []ProductSummary       `json:"recommended_products"`
	TotalProducts         int                    `json:"total_products"`
	AIProvider            string                 `json:"ai_provider"`
	ProcessingTimeMs      int64                  `json:"processing_time_ms"`
}

type ProductSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PointCost   int    `json:"point_cost"`
	ImageURL    string `json:"image_url"`
	InStock     bool   `json:"in_stock"`
	StockCount  int    `json:"stock_count"`
}

type ErrorDetails struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Details     string `json:"details,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Error      *GeminiError      `json:"error,omitempty"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type SimpleAIResponse struct {
	RecommendedCategoryID string `json:"recommended_category_id"`
	MinPointsLLM         int    `json:"min_points_llm"`
	MaxPointsLLM         int    `json:"max_points_llm"`
	Reasoning            string `json:"reasoning"`
}

func NewAIHandler(productService *services.ProductService, userRepo *repositories.UserRepository, categoryRepo *repositories.CategoryRepository) *AIHandler {
	return &AIHandler{
		ProductService: productService,
		UserRepo:       userRepo,
		CategoryRepo:   categoryRepo,
		Logger:         logger.NewZapLogger(),
	}
}

func (h *AIHandler) GetRecommendation(c *gin.Context) {
	startTime := time.Now()
	requestID := generateRequestID()
	
	h.Logger.Info("AI recommendation request started",
		zap.String("request_id", requestID),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	var req AIRecommendationRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Error("Invalid request body",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "INVALID_REQUEST", "Invalid request format", err.Error(), "Please provide a valid user_id in the request body")
		return
	}
	
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.Logger.Error("Invalid user ID format",
			zap.String("request_id", requestID),
			zap.String("user_id", req.UserID),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "INVALID_USER_ID", "Invalid user ID format", err.Error(), "Please provide a valid UUID format for user_id")
		return
	}
	
	// Get user's point balance
	user, err := h.UserRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.Logger.Error("Failed to fetch user",
			zap.String("request_id", requestID),
			zap.String("user_id", req.UserID),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "USER_NOT_FOUND", "User not found", err.Error(), "Please verify the user_id exists in the system")
		return
	}
	
	h.Logger.Info("User found",
		zap.String("request_id", requestID),
		zap.String("user_id", req.UserID),
		zap.String("user_name", user.Name),
		zap.Int("point_balance", user.PointBalance),
	)
	
	// Get all categories
	categories, err := h.CategoryRepo.GetAll(c.Request.Context())
	if err != nil {
		h.Logger.Error("Failed to fetch categories",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "CATEGORIES_FETCH_ERROR", "Failed to fetch categories", err.Error(), "Please try again later or contact support")
		return
	}
	
	if len(categories) == 0 {
		h.Logger.Warn("No categories available",
			zap.String("request_id", requestID),
		)
		h.sendErrorResponse(c, requestID, "NO_CATEGORIES", "No categories available", "No product categories found in the system", "Please contact admin to add product categories")
		return
	}
	
	h.Logger.Info("Categories fetched",
		zap.String("request_id", requestID),
		zap.Int("category_count", len(categories)),
	)
	
	// Get AI recommendation
	aiRecommendation, aiProvider, err := h.getGeminiRecommendation(c.Request.Context(), requestID, user.PointBalance, categories)
	if err != nil {
		h.Logger.Warn("Gemini AI failed, falling back to simple recommendation",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		// Fallback to simple recommendation if AI fails
		aiRecommendation = h.getSimpleRecommendation(user.PointBalance, categories)
		aiProvider = "fallback_rules"
	}
	
	// Parse recommended category ID
	recommendedCategoryID, err := uuid.Parse(aiRecommendation.RecommendedCategoryID)
	if err != nil {
		h.Logger.Error("Invalid category ID from recommendation",
			zap.String("request_id", requestID),
			zap.String("category_id", aiRecommendation.RecommendedCategoryID),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "INVALID_CATEGORY_ID", "Invalid category ID from recommendation", err.Error(), "Please try again or contact support")
		return
	}
	
	// Get category details
	category, err := h.CategoryRepo.GetByID(c.Request.Context(), recommendedCategoryID)
	if err != nil {
		h.Logger.Error("Failed to fetch category details",
			zap.String("request_id", requestID),
			zap.String("category_id", recommendedCategoryID.String()),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "CATEGORY_NOT_FOUND", "Recommended category not found", err.Error(), "Please try again or contact support")
		return
	}
	
	h.Logger.Info("Category details fetched",
		zap.String("request_id", requestID),
		zap.String("category_id", recommendedCategoryID.String()),
		zap.String("category_name", category.Name),
	)
	
	// Get products in the recommended category
	products, totalProducts, err := h.ProductService.GetProductsByCategory(c.Request.Context(), recommendedCategoryID, 1, 20)
	if err != nil {
		h.Logger.Error("Failed to fetch category products",
			zap.String("request_id", requestID),
			zap.String("category_id", recommendedCategoryID.String()),
			zap.Error(err),
		)
		h.sendErrorResponse(c, requestID, "PRODUCTS_FETCH_ERROR", "Failed to fetch category products", err.Error(), "Please try again later")
		return
	}
	
	h.Logger.Info("Products fetched",
		zap.String("request_id", requestID),
		zap.String("category_id", recommendedCategoryID.String()),
		zap.Int("product_count", len(products)),
		zap.Int("total_products", totalProducts),
	)
	
	// Convert products to summary format
	var productSummaries []ProductSummary
	for _, product := range products {
		productSummaries = append(productSummaries, ProductSummary{
			ID:          product.ID.String(),
			Name:        product.Name,
			Description: product.Description,
			PointCost:   product.PointCost,
			ImageURL:    product.ImageURL,
			InStock:     product.StockQuantity > 0,
			StockCount:  product.StockQuantity,
		})
	}
	
	processingTime := time.Since(startTime).Milliseconds()
	
	// Build complete response
	response := AIRecommendationResponse{
		Success:   true,
		Message:   "AI recommendation generated successfully",
		Timestamp: time.Now(),
		RequestID: requestID,
		Data: &RecommendationData{
			UserID:                req.UserID,
			PointBalance:          user.PointBalance,
			RecommendedCategoryID: aiRecommendation.RecommendedCategoryID,
			CategoryName:          category.Name,
			MinPointsLLM:          aiRecommendation.MinPointsLLM,
			MaxPointsLLM:          aiRecommendation.MaxPointsLLM,
			Reasoning:             aiRecommendation.Reasoning,
			RecommendedProducts:   productSummaries,
			TotalProducts:         totalProducts,
			AIProvider:            aiProvider,
			ProcessingTimeMs:      processingTime,
		},
	}
	
	h.Logger.Info("AI recommendation completed successfully",
		zap.String("request_id", requestID),
		zap.String("user_id", req.UserID),
		zap.String("recommended_category", category.Name),
		zap.Int("recommended_products", len(productSummaries)),
		zap.String("ai_provider", aiProvider),
		zap.Int64("processing_time_ms", processingTime),
	)
	
	c.JSON(http.StatusOK, response)
}

func (h *AIHandler) getGeminiRecommendation(ctx context.Context, requestID string, pointBalance int, categories []interface{}) (*SimpleAIResponse, string, error) {
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		h.Logger.Error("Gemini API key not configured", zap.String("request_id", requestID))
		return nil, "", fmt.Errorf("Gemini API key not configured")
	}
	
	// Prepare categories for prompt
	categoriesJSON, err := json.Marshal(categories)
	if err != nil {
		h.Logger.Error("Failed to marshal categories", zap.String("request_id", requestID), zap.Error(err))
		return nil, "", err
	}
	
	prompt := fmt.Sprintf(`Given a user with a current point balance of %d and the following available categories (each with an id and name): %s, suggest the most suitable category by its id and a corresponding minimum and maximum point range for products within that category. 

Consider the user's point balance and recommend a category that offers good value. The point range should be realistic for products in that category.

Ensure the response is a JSON object with the following fields: {"recommended_category_id": "string", "min_points_llm": "integer", "max_points_llm": "integer", "reasoning": "string"}.

Only return the JSON object, no additional text.`, 
		pointBalance, string(categoriesJSON))
	
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{
						Text: prompt,
					},
				},
			},
		},
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		h.Logger.Error("Failed to marshal Gemini request", zap.String("request_id", requestID), zap.Error(err))
		return nil, "", err
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=%s", geminiKey)
	
	h.Logger.Info("Sending request to Gemini AI",
		zap.String("request_id", requestID),
		zap.Int("point_balance", pointBalance),
		zap.Int("categories_count", len(categories)),
	)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		h.Logger.Error("Failed to create Gemini request", zap.String("request_id", requestID), zap.Error(err))
		return nil, "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		h.Logger.Error("Gemini API request failed", zap.String("request_id", requestID), zap.Error(err))
		return nil, "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.Logger.Error("Failed to read Gemini response", zap.String("request_id", requestID), zap.Error(err))
		return nil, "", err
	}
	
	if resp.StatusCode != http.StatusOK {
		h.Logger.Error("Gemini API error",
			zap.String("request_id", requestID),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", string(body)),
		)
		return nil, "", fmt.Errorf("Gemini API error: %s", string(body))
	}
	
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		h.Logger.Error("Failed to unmarshal Gemini response", zap.String("request_id", requestID), zap.Error(err))
		return nil, "", err
	}
	
	if geminiResp.Error != nil {
		h.Logger.Error("Gemini API returned error",
			zap.String("request_id", requestID),
			zap.Int("error_code", geminiResp.Error.Code),
			zap.String("error_message", geminiResp.Error.Message),
		)
		return nil, "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}
	
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		h.Logger.Error("Empty response from Gemini", zap.String("request_id", requestID))
		return nil, "", fmt.Errorf("no response from Gemini")
	}
	
	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	
	h.Logger.Info("Received response from Gemini",
		zap.String("request_id", requestID),
		zap.String("response_text", responseText),
	)
	
	var recommendation SimpleAIResponse
	if err := json.Unmarshal([]byte(responseText), &recommendation); err != nil {
		h.Logger.Error("Failed to parse Gemini response JSON",
			zap.String("request_id", requestID),
			zap.String("response_text", responseText),
			zap.Error(err),
		)
		return nil, "", fmt.Errorf("failed to parse Gemini response: %v", err)
	}
	
	h.Logger.Info("Successfully parsed Gemini recommendation",
		zap.String("request_id", requestID),
		zap.String("recommended_category_id", recommendation.RecommendedCategoryID),
		zap.Int("min_points", recommendation.MinPointsLLM),
		zap.Int("max_points", recommendation.MaxPointsLLM),
	)
	
	return &recommendation, "gemini-1.5-flash", nil
}

// Simple recommendation logic (fallback when AI is not available)
func (h *AIHandler) getSimpleRecommendation(pointBalance int, categories []interface{}) *SimpleAIResponse {
	// Default to first category if available
	var categoryID string
	if len(categories) > 0 {
		if cat, ok := categories[0].(map[string]interface{}); ok {
			if id, exists := cat["id"]; exists {
				categoryID = fmt.Sprintf("%v", id)
			}
		}
	}
	
	if pointBalance >= 1000 {
		return &SimpleAIResponse{
			RecommendedCategoryID: categoryID,
			MinPointsLLM:         500,
			MaxPointsLLM:         1500,
			Reasoning:            "You have enough points for premium products! Consider high-value electronics or gadgets.",
		}
	} else if pointBalance >= 500 {
		return &SimpleAIResponse{
			RecommendedCategoryID: categoryID,
			MinPointsLLM:         200,
			MaxPointsLLM:         600,
			Reasoning:            "Great! You can redeem mid-range products like accessories or books.",
		}
	} else if pointBalance >= 100 {
		return &SimpleAIResponse{
			RecommendedCategoryID: categoryID,
			MinPointsLLM:         50,
			MaxPointsLLM:         150,
			Reasoning:            "You can redeem small items or save up for bigger rewards!",
		}
	} else {
		return &SimpleAIResponse{
			RecommendedCategoryID: categoryID,
			MinPointsLLM:         0,
			MaxPointsLLM:         100,
			Reasoning:            "Keep earning points! Consider purchasing more credit packages.",
		}
	}
}

func (h *AIHandler) sendErrorResponse(c *gin.Context, requestID, code, message, details, suggestion string) {
	response := AIRecommendationResponse{
		Success:   false,
		Message:   "AI recommendation failed",
		Timestamp: time.Now(),
		RequestID: requestID,
		Error: &ErrorDetails{
			Code:       code,
			Message:    message,
			Details:    details,
			Suggestion: suggestion,
		},
	}
	
	statusCode := http.StatusInternalServerError
	switch code {
	case "INVALID_REQUEST", "INVALID_USER_ID":
		statusCode = http.StatusBadRequest
	case "USER_NOT_FOUND", "CATEGORY_NOT_FOUND":
		statusCode = http.StatusNotFound
	case "NO_CATEGORIES":
		statusCode = http.StatusServiceUnavailable
	}
	
	c.JSON(statusCode, response)
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d_%s", time.Now().Unix(), uuid.New().String()[:8])
}