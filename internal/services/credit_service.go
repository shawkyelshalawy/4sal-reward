package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
	"github.com/shawkyelshalawy/4sal-reward/internal/repositories"
)

type CreditService struct {
	CreditRepo *repositories.CreditRepository
	UserRepo   *repositories.UserRepository
}

func (s *CreditService) PurchasePackage(ctx context.Context, userID, packageID uuid.UUID, amountPaid float64) error {
	return s.CreditRepo.PurchasePackage(ctx, userID, packageID, amountPaid)
}

func (s *CreditService) GetPackage(ctx context.Context, packageID uuid.UUID) (*models.CreditPackage, error) {
	return s.CreditRepo.GetPackage(ctx, packageID)
}

func (s *CreditService) CreatePackage(ctx context.Context, name, description string, price float64, rewardPoints int) (uuid.UUID, error) {
	pkg := &models.CreditPackage{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		Price:        price,
		RewardPoints: rewardPoints,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	err := s.CreditRepo.CreatePackage(ctx, pkg)
	return pkg.ID, err
}

func (s *CreditService) UpdatePackage(ctx context.Context, packageID uuid.UUID, name, description *string, price *float64, rewardPoints *int, isActive *bool) error {
	return s.CreditRepo.UpdatePackage(ctx, packageID, name, description, price, rewardPoints, isActive)
}

// GetPackages retrieves credit packages with pagination
func (s *CreditService) GetPackages(ctx context.Context, page, size int) ([]models.CreditPackage, int, error) {
	return s.CreditRepo.GetPackages(ctx, page, size)
}