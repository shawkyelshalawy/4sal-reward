package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
	"github.com/stretchr/testify/mock"
)

type MockCreditRepository struct {
	mock.Mock
}

func (m *MockCreditRepository) PurchasePackage(ctx context.Context, userID, packageID uuid.UUID, amountPaid float64) error {
	args := m.Called(ctx, userID, packageID, amountPaid)
	return args.Error(0)
}

func (m *MockCreditRepository) GetPackage(ctx context.Context, packageID uuid.UUID) (*models.CreditPackage, error) {
	args := m.Called(ctx, packageID)
	return args.Get(0).(*models.CreditPackage), args.Error(1)
}

func (m *MockCreditRepository) CreatePackage(ctx context.Context, pkg *models.CreditPackage) error {
	args := m.Called(ctx, pkg)
	return args.Error(0)
}
