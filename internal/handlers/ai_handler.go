package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/repositories"
	"github.com/shawkyelshalawy/4sal-reward/internal/services"
)

type AIHandler struct {
	ProductService *services.ProductService
	UserRepo       *repositories.UserRepository
	CategoryRepo   *repositories.CategoryRepository
}

type AIRecommendationRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type AIRecommendationResponse struct {
	RecommendedCategoryID string `json:"recommended_category_id"`
	MinPointsLLM         int    `json:"min_points_llm"`
	MaxPointsLLM         int    `json:"max_points_llm"`
	Reasoning            string `json:"reasoning"`
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
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

func (h *AIHandler) GetRecommendation(c *gin.Context) {
	var req AIRecommendationRequest
	
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
	
	// Get all categories
	categories, err := h.CategoryRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	
	// Get AI recommendation
	recommendation, err := h.getGeminiRecommendation(c.Request.Context(), user.PointBalance, categories)
	if err != nil {
		// Fallback to simple recommendation if AI fails
		recommendation = h.getSimpleRecommendation(user.PointBalance, categories)
	}
	
	c.JSON(http.StatusOK, recommendation)
}

func (h *AIHandler) getGeminiRecommendation(ctx context.Context, pointBalance int, categories []interface{}) (*AIRecommendationResponse, error) {
	geminiKey := "AIzaSyCCLOJCy5DwAUoSFgInnqbW7AkQJQyt_-Q"
	
	// Prepare categories for prompt
	categoriesJSON, _ := json.Marshal(categories)
	
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
		return nil, err
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=%s", geminiKey)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini API error: %s", string(body))
	}
	
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, err
	}
	
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}
	
	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	
	var recommendation AIRecommendationResponse
	if err := json.Unmarshal([]byte(responseText), &recommendation); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %v", err)
	}
	
	return &recommendation, nil
}

// Simple recommendation logic (fallback when AI is not available)
func (h *AIHandler) getSimpleRecommendation(pointBalance int, categories []interface{}) *AIRecommendationResponse {
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
		return &AIRecommendationResponse{
			RecommendedCategoryID: categoryID,
			MinPointsLLM:         500,
			MaxPointsLLM:         1500,
			Reasoning:            "You have enough points for premium products! Consider high-value electronics or gadgets.",
		}
	} else if pointBalance >= 500 {
		return &AIRecommendationResponse{
			RecommendedCategoryID: categoryID,
			MinPointsLLM:         200,
			MaxPointsLLM:         600,
			Reasoning:            "Great! You can redeem mid-range products like accessories or books.",
		}
	} else if pointBalance >= 100 {
		return &AIRecommendationResponse{
			RecommendedCategoryID: categoryID,
			MinPointsLLM:         50,
			MaxPointsLLM:         150,
			Reasoning:            "You can redeem small items or save up for bigger rewards!",
		}
	} else {
		return &AIRecommendationResponse{
			RecommendedCategoryID: categoryID,
			MinPointsLLM:         0,
			MaxPointsLLM:         100,
			Reasoning:            "Keep earning points! Consider purchasing more credit packages.",
		}
	}
}