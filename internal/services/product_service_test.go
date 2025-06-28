package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository implements the product repository interface for testing
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Search(ctx context.Context, query string, page, size int) ([]models.Product, error) {
	args := m.Called(ctx, query, page, size)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) UpdateOfferStatus(ctx context.Context, productID uuid.UUID, isInOfferPool bool) error {
	args := m.Called(ctx, productID, isInOfferPool)
	return args.Error(0)
}

func (m *MockProductRepository) RedeemProduct(ctx context.Context, userID, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, userID, productID, quantity)
	return args.Error(0)
}

func (m *MockProductRepository) UpdateProduct(ctx context.Context, productID uuid.UUID, name, description *string, categoryID *uuid.UUID, pointCost, stockQuantity *int, isActive, isInOfferPool *bool, imageURL *string) error {
	args := m.Called(ctx, productID, name, description, categoryID, pointCost, stockQuantity, isActive, isInOfferPool, imageURL)
	return args.Error(0)
}

func (m *MockProductRepository) GetProducts(ctx context.Context, page, size int, isActive, isInOfferPool, categoryID string) ([]models.Product, int, error) {
	args := m.Called(ctx, page, size, isActive, isInOfferPool, categoryID)
	return args.Get(0).([]models.Product), args.Int(1), args.Error(2)
}

func (m *MockProductRepository) GetProductsByCategory(ctx context.Context, categoryID uuid.UUID, page, size int) ([]models.Product, int, error) {
	args := m.Called(ctx, categoryID, page, size)
	return args.Get(0).([]models.Product), args.Int(1), args.Error(2)
}

// MockRedisClient implements the Redis client interface for testing
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	cmd := redis.NewStringCmd(ctx)
	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	} else {
		cmd.SetVal(args.String(0))
	}
	return cmd
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	cmd := redis.NewStatusCmd(ctx)
	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	}
	return cmd
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	cmd := redis.NewIntCmd(ctx)
	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	}
	return cmd
}

func (m *MockRedisClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	args := m.Called(ctx, pattern)
	cmd := redis.NewStringSliceCmd(ctx)
	if args.Error(0) != nil {
		cmd.SetErr(args.Error(0))
	} else {
		cmd.SetVal(args.Get(0).([]string))
	}
	return cmd
}

func TestProductService_SearchProducts_CacheHit(t *testing.T) {
	// Arrange
	mockProductRepo := new(MockProductRepository)
	mockRedisClient := new(MockRedisClient)
	service := &ProductService{
		ProductRepo: mockProductRepo,
		RedisClient: mockRedisClient,
	}

	ctx := context.Background()
	query := "wireless"
	page := 1
	size := 10
	cacheKey := "product:search:wireless:1:10"

	expectedProducts := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Wireless Earbuds",
			Description: "High-quality wireless earbuds",
			PointCost:   500,
		},
	}

	cachedData, _ := json.Marshal(expectedProducts)
	mockRedisClient.On("Get", ctx, cacheKey).Return(string(cachedData), nil)

	// Act
	products, err := service.SearchProducts(ctx, query, page, size)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, "Wireless Earbuds", products[0].Name)
	mockRedisClient.AssertExpectations(t)
	mockProductRepo.AssertNotCalled(t, "Search")
}

func TestProductService_SearchProducts_CacheMiss(t *testing.T) {
	// Arrange
	mockProductRepo := new(MockProductRepository)
	mockRedisClient := new(MockRedisClient)
	service := &ProductService{
		ProductRepo: mockProductRepo,
		RedisClient: mockRedisClient,
	}

	ctx := context.Background()
	query := "wireless"
	page := 1
	size := 10
	cacheKey := "product:search:wireless:1:10"

	expectedProducts := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Wireless Earbuds",
			Description: "High-quality wireless earbuds",
			PointCost:   500,
		},
	}

	mockRedisClient.On("Get", ctx, cacheKey).Return("", redis.Nil)
	mockProductRepo.On("Search", ctx, query, page, size).Return(expectedProducts, nil)
	mockRedisClient.On("Set", ctx, cacheKey, mock.AnythingOfType("[]uint8"), productSearchCacheTTL).Return(nil)

	// Act
	products, err := service.SearchProducts(ctx, query, page, size)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, "Wireless Earbuds", products[0].Name)
	mockRedisClient.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}

func TestProductService_SearchProducts_RepositoryError(t *testing.T) {
	// Arrange
	mockProductRepo := new(MockProductRepository)
	mockRedisClient := new(MockRedisClient)
	service := &ProductService{
		ProductRepo: mockProductRepo,
		RedisClient: mockRedisClient,
	}

	ctx := context.Background()
	query := "wireless"
	page := 1
	size := 10
	cacheKey := "product:search:wireless:1:10"
	expectedError := assert.AnError

	mockRedisClient.On("Get", ctx, cacheKey).Return("", redis.Nil)
	mockProductRepo.On("Search", ctx, query, page, size).Return([]models.Product{}, expectedError)

	// Act
	products, err := service.SearchProducts(ctx, query, page, size)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, products)
	mockRedisClient.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}

func TestProductService_RedeemProduct_Success(t *testing.T) {
	// Arrange
	mockProductRepo := new(MockProductRepository)
	mockRedisClient := new(MockRedisClient)
	service := &ProductService{
		ProductRepo: mockProductRepo,
		RedisClient: mockRedisClient,
	}

	ctx := context.Background()
	userID := uuid.New()
	productID := uuid.New()
	quantity := 2

	mockProductRepo.On("RedeemProduct", ctx, userID, productID, quantity).Return(nil)

	// Act
	err := service.RedeemProduct(ctx, userID, productID, quantity)

	// Assert
	assert.NoError(t, err)
	mockProductRepo.AssertExpectations(t)
}

func TestProductService_RedeemProduct_Error(t *testing.T) {
	// Arrange
	mockProductRepo := new(MockProductRepository)
	mockRedisClient := new(MockRedisClient)
	service := &ProductService{
		ProductRepo: mockProductRepo,
		RedisClient: mockRedisClient,
	}

	ctx := context.Background()
	userID := uuid.New()
	productID := uuid.New()
	quantity := 2
	expectedError := assert.AnError

	mockProductRepo.On("RedeemProduct", ctx, userID, productID, quantity).Return(expectedError)

	// Act
	err := service.RedeemProduct(ctx, userID, productID, quantity)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockProductRepo.AssertExpectations(t)
}

func TestProductService_CreateProduct_Success(t *testing.T) {
	// Arrange
	mockProductRepo := new(MockProductRepository)
	mockRedisClient := new(MockRedisClient)
	service := &ProductService{
		ProductRepo: mockProductRepo,
		RedisClient: mockRedisClient,
	}

	ctx := context.Background()
	name := "Test Product"
	description := "Test Description"
	categoryID := uuid.New()
	pointCost := 500
	stockQuantity := 10
	isInOfferPool := true
	imageURL := "https://example.com/image.jpg"

	mockProductRepo.On("Create", ctx, mock.MatchedBy(func(product *models.Product) bool {
		return product.Name == name &&
			product.Description == description &&
			product.CategoryID != nil &&
			*product.CategoryID == categoryID &&
			product.PointCost == pointCost &&
			product.StockQuantity == stockQuantity &&
			product.IsActive == true &&
			product.IsInOfferPool == isInOfferPool &&
			product.ImageURL == imageURL
	})).Return(nil)

	// Act
	productID, err := service.CreateProduct(ctx, name, description, &categoryID, pointCost, stockQuantity, isInOfferPool, imageURL)

	// Assert
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, productID)
	mockProductRepo.AssertExpectations(t)
}

func TestProductService_CreateProduct_Error(t *testing.T) {
	// Arrange
	mockProductRepo := new(MockProductRepository)
	mockRedisClient := new(MockRedisClient)
	service := &ProductService{
		ProductRepo: mockProductRepo,
		RedisClient: mockRedisClient,
	}

	ctx := context.Background()
	name := "Test Product"
	description := "Test Description"
	categoryID := uuid.New()
	pointCost := 500
	stockQuantity := 10
	isInOfferPool := true
	imageURL := "https://example.com/image.jpg"
	expectedError := assert.AnError

	mockProductRepo.On("Create", ctx, mock.AnythingOfType("*models.Product")).Return(expectedError)

	// Act
	productID, err := service.CreateProduct(ctx, name, description, &categoryID, pointCost, stockQuantity, isInOfferPool, imageURL)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, productID)
	assert.Equal(t, expectedError, err)
	mockProductRepo.AssertExpectations(t)
}

func TestProductService_UpdateOfferStatus_Success(t *testing.T) {
	// Arrange
	mockProductRepo := new(MockProductRepository)
	mockRedisClient := new(MockRedisClient)
	service := &ProductService{
		ProductRepo: mockProductRepo,
		RedisClient: mockRedisClient,
	}

	ctx := context.Background()
	productID := uuid.New()
	isInOfferPool := false

	mockProductRepo.On("UpdateOfferStatus", ctx, productID, isInOfferPool).Return(nil)

	// Act
	err := service.UpdateOfferStatus(ctx, productID, isInOfferPool)

	// Assert
	assert.NoError(t, err)
	mockProductRepo.AssertExpectations(t)
}

func TestProductService_GetProductsByCategory_Success(t *testing.T) {
	// Arrange
	mockProductRepo := new(MockProductRepository)
	mockRedisClient := new(MockRedisClient)
	service := &ProductService{
		ProductRepo: mockProductRepo,
		RedisClient: mockRedisClient,
	}

	ctx := context.Background()
	categoryID := uuid.New()
	page := 1
	size := 10

	expectedProducts := []models.Product{
		{
			ID:         uuid.New(),
			Name:       "Product 1",
			PointCost:  500,
			CategoryID: &categoryID,
		},
		{
			ID:         uuid.New(),
			Name:       "Product 2",
			PointCost:  800,
			CategoryID: &categoryID,
		},
	}
	expectedTotal := 2

	mockProductRepo.On("GetProductsByCategory", categoryID, page, size).Return(expectedProducts, expectedTotal, nil)

	// Act
	products, total, err := service.GetProductsByCategory(ctx, categoryID, page, size)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedProducts, products)
	assert.Equal(t, expectedTotal, total)
	mockProductRepo.AssertExpectations(t)
}

func TestProductService_UpdateProduct_WithCacheClear(t *testing.T) {
	// Arrange
	mockProductRepo := new(MockProductRepository)
	mockRedisClient := new(MockRedisClient)
	service := &ProductService{
		ProductRepo: mockProductRepo,
		RedisClient: mockRedisClient,
	}

	ctx := context.Background()
	productID := uuid.New()
	name := "Updated Product"
	stockQuantity := 15

	// Mock cache clearing
	mockRedisClient.On("Keys", ctx, "product:search:*").Return([]string{"product:search:key1", "product:search:key2"}, nil)
	mockRedisClient.On("Del", ctx, []string{"product:search:key1", "product:search:key2"}).Return(nil)
	mockProductRepo.On("UpdateProduct", ctx, productID, &name, (*string)(nil), (*uuid.UUID)(nil), (*int)(nil), &stockQuantity, (*bool)(nil), (*bool)(nil), (*string)(nil)).Return(nil)

	// Act
	err := service.UpdateProduct(ctx, productID, &name, nil, nil, nil, &stockQuantity, nil, nil, nil)

	// Assert
	assert.NoError(t, err)
	mockRedisClient.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}