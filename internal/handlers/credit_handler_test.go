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

// MockCreditService implements the credit service interface for testing
type MockCreditService struct {
	mock.Mock
}

func (m *MockCreditService) PurchasePackage(ctx context.Context, userID, packageID uuid.UUID, amountPaid float64) error {
	args := m.Called(ctx, userID, packageID, amountPaid)
	return args.Error(0)
}

func (m *MockCreditService) GetPackage(ctx context.Context, packageID uuid.UUID) (*models.CreditPackage, error) {
	args := m.Called(ctx, packageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CreditPackage), args.Error(1)
}

func (m *MockCreditService) CreatePackage(ctx context.Context, name, description string, price float64, rewardPoints int) (uuid.UUID, error) {
	args := m.Called(ctx, name, description, price, rewardPoints)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockCreditService) UpdatePackage(ctx context.Context, packageID uuid.UUID, name, description *string, price *float64, rewardPoints *int, isActive *bool) error {
	args := m.Called(ctx, packageID, name, description, price, rewardPoints, isActive)
	return args.Error(0)
}

func (m *MockCreditService) GetPackages(ctx context.Context, page, size int) ([]models.CreditPackage, int, error) {
	args := m.Called(ctx, page, size)
	return args.Get(0).([]models.CreditPackage), args.Int(1), args.Error(2)
}

func setupCreditHandler() (*CreditHandler, *MockCreditService) {
	mockService := new(MockCreditService)
	handler := NewCreditHandler(mockService)
	return handler, mockService
}

func TestCreditHandler_PurchaseCreditPackage_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCreditHandler()

	userID := uuid.New()
	packageID := uuid.New()
	amountPaid := 50.0

	requestBody := map[string]interface{}{
		"user_id":     userID.String(),
		"package_id":  packageID.String(),
		"amount_paid": amountPaid,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/credits/purchase", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("PurchasePackage", mock.AnythingOfType("*context.emptyCtx"), userID, packageID, amountPaid).Return(nil)

	// Act
	handler.PurchaseCreditPackage(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Credit package purchased successfully", response.Message)
	assert.NotEmpty(t, response.RequestID)

	mockService.AssertExpectations(t)
}

func TestCreditHandler_PurchaseCreditPackage_InvalidJSON(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, _ := setupCreditHandler()

	req, _ := http.NewRequest("POST", "/credits/purchase", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	handler.PurchaseCreditPackage(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_REQUEST", response.Error.Code)
}

func TestCreditHandler_PurchaseCreditPackage_InvalidUserID(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, _ := setupCreditHandler()

	requestBody := map[string]interface{}{
		"user_id":     "invalid-uuid",
		"package_id":  uuid.New().String(),
		"amount_paid": 50.0,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/credits/purchase", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	handler.PurchaseCreditPackage(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_USER_ID", response.Error.Code)
}

func TestCreditHandler_PurchaseCreditPackage_InvalidAmount(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, _ := setupCreditHandler()

	requestBody := map[string]interface{}{
		"user_id":     uuid.New().String(),
		"package_id":  uuid.New().String(),
		"amount_paid": -10.0,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/credits/purchase", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Act
	handler.PurchaseCreditPackage(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_AMOUNT", response.Error.Code)
}

func TestCreditHandler_PurchaseCreditPackage_ServiceError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCreditHandler()

	userID := uuid.New()
	packageID := uuid.New()
	amountPaid := 50.0

	requestBody := map[string]interface{}{
		"user_id":     userID.String(),
		"package_id":  packageID.String(),
		"amount_paid": amountPaid,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/credits/purchase", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("PurchasePackage", mock.AnythingOfType("*context.emptyCtx"), userID, packageID, amountPaid).Return(assert.AnError)

	// Act
	handler.PurchaseCreditPackage(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "PURCHASE_FAILED", response.Error.Code)

	mockService.AssertExpectations(t)
}

func TestCreditHandler_GetCreditPackages_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCreditHandler()

	expectedPackages := []models.CreditPackage{
		{
			ID:           uuid.New(),
			Name:         "Bronze Package",
			Price:        25.0,
			RewardPoints: 250,
			IsActive:     true,
		},
		{
			ID:           uuid.New(),
			Name:         "Silver Package",
			Price:        50.0,
			RewardPoints: 500,
			IsActive:     true,
		},
	}
	expectedTotal := 2

	req, _ := http.NewRequest("GET", "/credits/packages?page=1&size=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("GetPackages", mock.AnythingOfType("*context.emptyCtx"), 1, 10).Return(expectedPackages, expectedTotal, nil)

	// Act
	handler.GetCreditPackages(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Credit packages retrieved successfully", response.Message)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(10), data["size"])
	assert.Equal(t, float64(2), data["total"])
	assert.Equal(t, float64(1), data["total_pages"])

	mockService.AssertExpectations(t)
}

func TestCreditHandler_GetCreditPackages_DefaultPagination(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCreditHandler()

	expectedPackages := []models.CreditPackage{}
	expectedTotal := 0

	req, _ := http.NewRequest("GET", "/credits/packages", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("GetPackages", mock.AnythingOfType("*context.emptyCtx"), 1, 10).Return(expectedPackages, expectedTotal, nil)

	// Act
	handler.GetCreditPackages(c)

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

func TestCreditHandler_GetCreditPackages_ServiceError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCreditHandler()

	req, _ := http.NewRequest("GET", "/credits/packages", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("GetPackages", mock.AnythingOfType("*context.emptyCtx"), 1, 10).Return([]models.CreditPackage{}, 0, assert.AnError)

	// Act
	handler.GetCreditPackages(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "PACKAGES_FETCH_ERROR", response.Error.Code)

	mockService.AssertExpectations(t)
}

func TestCreditHandler_GetCreditPackage_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCreditHandler()

	packageID := uuid.New()
	expectedPackage := &models.CreditPackage{
		ID:           packageID,
		Name:         "Test Package",
		Price:        50.0,
		RewardPoints: 500,
		IsActive:     true,
	}

	req, _ := http.NewRequest("GET", "/credits/packages/"+packageID.String(), nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: packageID.String()}}

	mockService.On("GetPackage", mock.AnythingOfType("*context.emptyCtx"), packageID).Return(expectedPackage, nil)

	// Act
	handler.GetCreditPackage(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Credit package retrieved successfully", response.Message)

	mockService.AssertExpectations(t)
}

func TestCreditHandler_GetCreditPackage_InvalidID(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, _ := setupCreditHandler()

	req, _ := http.NewRequest("GET", "/credits/packages/invalid-uuid", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	// Act
	handler.GetCreditPackage(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_PACKAGE_ID", response.Error.Code)
}

func TestCreditHandler_GetCreditPackage_NotFound(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	handler, mockService := setupCreditHandler()

	packageID := uuid.New()

	req, _ := http.NewRequest("GET", "/credits/packages/"+packageID.String(), nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: packageID.String()}}

	mockService.On("GetPackage", mock.AnythingOfType("*context.emptyCtx"), packageID).Return(nil, assert.AnError)

	// Act
	handler.GetCreditPackage(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "PACKAGE_NOT_FOUND", response.Error.Code)

	mockService.AssertExpectations(t)
}