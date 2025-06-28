package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductService implements the product service interface for testing
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) SearchProducts(ctx context.Context, query string, page, size int) ([]models.Product, error) {
	args := m.Called(ctx, query, page, size)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductService) RedeemProduct(ctx context.Context, userID, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, userID, productID, quantity)
	return args.Error(0)
}

func (m *MockProductService) CreateProduct(ctx context.Context, name, description string, categoryID *uuid.UUID, pointCost, stockQuantity int, isInOfferPool bool, imageURL string) (uuid.UUID, error) {
	args := m.Called(ctx, name, description, categoryID, pointCost, stockQuantity, isInOfferPool, imageURL)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockProductService) UpdateOfferStatus(ctx context.Context, productID uuid.UUID, isInOfferPool bool) error {
	args := m.Called(ctx, productID, isInOfferPool)
	return args.Error(0)
}

func (m *MockProductService) UpdateProduct(ctx context.Context, productID uuid.UUID, name, description *string, categoryID *uuid.UUID, pointCost, stockQuantity *int, isActive, isInOfferPool *bool, imageURL *string) error {
	args := m.Called(ctx, productID, name, description, categoryID, pointCost, stockQuantity, isActive, isInOfferPool, imageURL)
	return args.Error(0)
}

func (m *MockProductService) GetProducts(ctx context.Context, page, size int, isActive, isInOfferPool, categoryID string) ([]models.Product, int, error) {
	args := m.Called(ctx, page, size, isActive, isInOfferPool, categoryID)
	return args.Get(0).([]models.Product), args.Int(1), args.Error(2)
}

func (m *MockProductService) GetProductsByCategory(ctx context.Context, categoryID uuid.UUID, page, size int) ([]models.Product, int, error) {
	args := m.Called(ctx, categoryID, page, size)
	return args.Get(0).([]models.Product), args.Int(1), args.Error(2)
}

func setupProductHandler() (*ProductHandler, *MockProductService) {
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)
	return handler, mockService
}

func TestProductHandler_SearchProducts_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupProductHandler()

	expectedProducts := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Wireless Earbuds",
			Description: "High-quality wireless earbuds",
			PointCost:   500,
			ImageURL:    "https://example.com/earbuds.jpg",
		},
		{
			ID:          uuid.New(),
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse",
			PointCost:   300,
			ImageURL:    "https://example.com/mouse.jpg",
		},
	}

	req, _ := http.NewRequest("GET", "/products/search?query=wireless&page=1&size=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("SearchProducts", mock.AnythingOfType("*context.emptyCtx"), "wireless", 1, 10).Return(expectedProducts, nil)

	// Act
	handler.SearchProducts(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Product search completed successfully", response.Message)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, "wireless", data["query"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(10), data["size"])

	mockService.AssertExpectations(t)
}

func TestProductHandler_SearchProducts_MissingQuery(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, _ := setupProductHandler()

	req, _ := http.NewRequest("GET", "/products/search", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	handler.SearchProducts(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "MISSING_QUERY", response.Error.Code)
}

func TestProductHandler_SearchProducts_DefaultPagination(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupProductHandler()

	expectedProducts := []models.Product{}

	req, _ := http.NewRequest("GET", "/products/search?query=test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("SearchProducts", mock.AnythingOfType("*context.emptyCtx"), "test", 1, 10).Return(expectedProducts, nil)

	// Act
	handler.SearchProducts(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(10), data["size"])

	mockService.AssertExpectations(t)
}

func TestProductHandler_SearchProducts_ServiceError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupProductHandler()

	req, _ := http.NewRequest("GET", "/products/search?query=wireless", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("SearchProducts", mock.AnythingOfType("*context.emptyCtx"), "wireless", 1, 10).Return([]models.Product{}, assert.AnError)

	// Act
	handler.SearchProducts(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "SEARCH_FAILED", response.Error.Code)

	mockService.AssertExpectations(t)
}

func TestProductHandler_RedeemProduct_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupProductHandler()

	userID := uuid.New()
	productID := uuid.New()
	quantity := 2

	requestBody := map[string]interface{}{
		"user_id":    userID.String(),
		"product_id": productID.String(),
		"quantity":   quantity,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/products/redeem", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("RedeemProduct", mock.AnythingOfType("*context.emptyCtx"), userID, productID, quantity).Return(nil)

	// Act
	handler.RedeemProduct(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Product redeemed successfully", response.Message)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, userID.String(), data["user_id"])
	assert.Equal(t, productID.String(), data["product_id"])
	assert.Equal(t, float64(quantity), data["quantity"])

	mockService.AssertExpectations(t)
}

func TestProductHandler_RedeemProduct_InvalidJSON(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, _ := setupProductHandler()

	req, _ := http.NewRequest("POST", "/products/redeem", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	handler.RedeemProduct(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_REQUEST", response.Error.Code)
}

func TestProductHandler_RedeemProduct_InvalidUserID(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, _ := setupProductHandler()

	requestBody := map[string]interface{}{
		"user_id":    "invalid-uuid",
		"product_id": uuid.New().String(),
		"quantity":   1,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/products/redeem", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	handler.RedeemProduct(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_USER_ID", response.Error.Code)
}

func TestProductHandler_RedeemProduct_InvalidProductID(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, _ := setupProductHandler()

	requestBody := map[string]interface{}{
		"user_id":    uuid.New().String(),
		"product_id": "invalid-uuid",
		"quantity":   1,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/products/redeem", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	handler.RedeemProduct(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_PRODUCT_ID", response.Error.Code)
}

func TestProductHandler_RedeemProduct_InvalidQuantity(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, _ := setupProductHandler()

	requestBody := map[string]interface{}{
		"user_id":    uuid.New().String(),
		"product_id": uuid.New().String(),
		"quantity":   0, // Invalid quantity
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/products/redeem", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	handler.RedeemProduct(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_REQUEST", response.Error.Code)
}

func TestProductHandler_RedeemProduct_ServiceError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupProductHandler()

	userID := uuid.New()
	productID := uuid.New()
	quantity := 1

	requestBody := map[string]interface{}{
		"user_id":    userID.String(),
		"product_id": productID.String(),
		"quantity":   quantity,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/products/redeem", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("RedeemProduct", mock.AnythingOfType("*context.emptyCtx"), userID, productID, quantity).Return(assert.AnError)

	// Act
	handler.RedeemProduct(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "REDEMPTION_FAILED", response.Error.Code)

	mockService.AssertExpectations(t)
}