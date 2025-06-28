package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCreditRepository implements the credit repository interface for testing
type MockCreditRepository struct {
	mock.Mock
}

func (m *MockCreditRepository) PurchasePackage(ctx context.Context, userID, packageID uuid.UUID, amountPaid float64) error {
	args := m.Called(ctx, userID, packageID, amountPaid)
	return args.Error(0)
}

func (m *MockCreditRepository) GetPackage(ctx context.Context, packageID uuid.UUID) (*models.CreditPackage, error) {
	args := m.Called(ctx, packageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CreditPackage), args.Error(1)
}

func (m *MockCreditRepository) CreatePackage(ctx context.Context, pkg *models.CreditPackage) error {
	args := m.Called(ctx, pkg)
	return args.Error(0)
}

func (m *MockCreditRepository) UpdatePackage(ctx context.Context, packageID uuid.UUID, name, description *string, price *float64, rewardPoints *int, isActive *bool) error {
	args := m.Called(ctx, packageID, name, description, price, rewardPoints, isActive)
	return args.Error(0)
}

func (m *MockCreditRepository) GetPackages(ctx context.Context, page, size int) ([]models.CreditPackage, int, error) {
	args := m.Called(ctx, page, size)
	return args.Get(0).([]models.CreditPackage), args.Int(1), args.Error(2)
}

// MockUserRepository implements the user repository interface for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) AddPoints(ctx context.Context, userID uuid.UUID, points int) error {
	args := m.Called(ctx, userID, points)
	return args.Error(0)
}

func (m *MockUserRepository) DeductPoints(ctx context.Context, userID uuid.UUID, points int) error {
	args := m.Called(ctx, userID, points)
	return args.Error(0)
}

func TestCreditService_PurchasePackage_Success(t *testing.T) {
	// Arrange
	mockCreditRepo := new(MockCreditRepository)
	mockUserRepo := new(MockUserRepository)
	service := &CreditService{
		CreditRepo: mockCreditRepo,
		UserRepo:   mockUserRepo,
	}

	ctx := context.Background()
	userID := uuid.New()
	packageID := uuid.New()
	amountPaid := 50.0

	mockCreditRepo.On("PurchasePackage", ctx, userID, packageID, amountPaid).Return(nil)

	// Act
	err := service.PurchasePackage(ctx, userID, packageID, amountPaid)

	// Assert
	assert.NoError(t, err)
	mockCreditRepo.AssertExpectations(t)
}

func TestCreditService_PurchasePackage_Error(t *testing.T) {
	// Arrange
	mockCreditRepo := new(MockCreditRepository)
	mockUserRepo := new(MockUserRepository)
	service := &CreditService{
		CreditRepo: mockCreditRepo,
		UserRepo:   mockUserRepo,
	}

	ctx := context.Background()
	userID := uuid.New()
	packageID := uuid.New()
	amountPaid := 50.0
	expectedError := assert.AnError

	mockCreditRepo.On("PurchasePackage", ctx, userID, packageID, amountPaid).Return(expectedError)

	// Act
	err := service.PurchasePackage(ctx, userID, packageID, amountPaid)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockCreditRepo.AssertExpectations(t)
}

func TestCreditService_GetPackage_Success(t *testing.T) {
	// Arrange
	mockCreditRepo := new(MockCreditRepository)
	mockUserRepo := new(MockUserRepository)
	service := &CreditService{
		CreditRepo: mockCreditRepo,
		UserRepo:   mockUserRepo,
	}

	ctx := context.Background()
	packageID := uuid.New()
	expectedPackage := &models.CreditPackage{
		ID:           packageID,
		Name:         "Test Package",
		Description:  "Test Description",
		Price:        50.0,
		RewardPoints: 500,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockCreditRepo.On("GetPackage", ctx, packageID).Return(expectedPackage, nil)

	// Act
	result, err := service.GetPackage(ctx, packageID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedPackage, result)
	mockCreditRepo.AssertExpectations(t)
}

func TestCreditService_GetPackage_NotFound(t *testing.T) {
	// Arrange
	mockCreditRepo := new(MockCreditRepository)
	mockUserRepo := new(MockUserRepository)
	service := &CreditService{
		CreditRepo: mockCreditRepo,
		UserRepo:   mockUserRepo,
	}

	ctx := context.Background()
	packageID := uuid.New()
	expectedError := assert.AnError

	mockCreditRepo.On("GetPackage", ctx, packageID).Return(nil, expectedError)

	// Act
	result, err := service.GetPackage(ctx, packageID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockCreditRepo.AssertExpectations(t)
}

func TestCreditService_CreatePackage_Success(t *testing.T) {
	// Arrange
	mockCreditRepo := new(MockCreditRepository)
	mockUserRepo := new(MockUserRepository)
	service := &CreditService{
		CreditRepo: mockCreditRepo,
		UserRepo:   mockUserRepo,
	}

	ctx := context.Background()
	name := "Test Package"
	description := "Test Description"
	price := 50.0
	rewardPoints := 500

	mockCreditRepo.On("CreatePackage", ctx, mock.MatchedBy(func(pkg *models.CreditPackage) bool {
		return pkg.Name == name &&
			pkg.Description == description &&
			pkg.Price == price &&
			pkg.RewardPoints == rewardPoints &&
			pkg.IsActive == true
	})).Return(nil)

	// Act
	packageID, err := service.CreatePackage(ctx, name, description, price, rewardPoints)

	// Assert
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, packageID)
	mockCreditRepo.AssertExpectations(t)
}

func TestCreditService_CreatePackage_Error(t *testing.T) {
	// Arrange
	mockCreditRepo := new(MockCreditRepository)
	mockUserRepo := new(MockUserRepository)
	service := &CreditService{
		CreditRepo: mockCreditRepo,
		UserRepo:   mockUserRepo,
	}

	ctx := context.Background()
	name := "Test Package"
	description := "Test Description"
	price := 50.0
	rewardPoints := 500
	expectedError := assert.AnError

	mockCreditRepo.On("CreatePackage", ctx, mock.AnythingOfType("*models.CreditPackage")).Return(expectedError)

	// Act
	packageID, err := service.CreatePackage(ctx, name, description, price, rewardPoints)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, packageID)
	assert.Equal(t, expectedError, err)
	mockCreditRepo.AssertExpectations(t)
}

func TestCreditService_UpdatePackage_Success(t *testing.T) {
	// Arrange
	mockCreditRepo := new(MockCreditRepository)
	mockUserRepo := new(MockUserRepository)
	service := &CreditService{
		CreditRepo: mockCreditRepo,
		UserRepo:   mockUserRepo,
	}

	ctx := context.Background()
	packageID := uuid.New()
	name := "Updated Package"
	price := 75.0
	isActive := false

	mockCreditRepo.On("UpdatePackage", ctx, packageID, &name, (*string)(nil), &price, (*int)(nil), &isActive).Return(nil)

	// Act
	err := service.UpdatePackage(ctx, packageID, &name, nil, &price, nil, &isActive)

	// Assert
	assert.NoError(t, err)
	mockCreditRepo.AssertExpectations(t)
}

func TestCreditService_GetPackages_Success(t *testing.T) {
	// Arrange
	mockCreditRepo := new(MockCreditRepository)
	mockUserRepo := new(MockUserRepository)
	service := &CreditService{
		CreditRepo: mockCreditRepo,
		UserRepo:   mockUserRepo,
	}

	ctx := context.Background()
	page := 1
	size := 10
	expectedPackages := []models.CreditPackage{
		{
			ID:           uuid.New(),
			Name:         "Package 1",
			Price:        25.0,
			RewardPoints: 250,
			IsActive:     true,
		},
		{
			ID:           uuid.New(),
			Name:         "Package 2",
			Price:        50.0,
			RewardPoints: 500,
			IsActive:     true,
		},
	}
	expectedTotal := 2

	mockCreditRepo.On("GetPackages", ctx, page, size).Return(expectedPackages, expectedTotal, nil)

	// Act
	packages, total, err := service.GetPackages(ctx, page, size)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedPackages, packages)
	assert.Equal(t, expectedTotal, total)
	mockCreditRepo.AssertExpectations(t)
}