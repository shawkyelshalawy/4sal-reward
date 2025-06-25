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

func (s *CreditService) PurchasePackage(ctx context.Context, userID, packageID uuid.UUID) error {
	pkg, err := s.CreditRepo.GetPackage(ctx, packageID)
	if err != nil {
		return err
	}
	
	if err := s.UserRepo.AddPoints(ctx, userID, pkg.RewardPoints); err != nil {
		return err
	}
	
	purchase := &models.CreditPurchase{
		ID:            uuid.New(),
		UserID:        userID,
		CreditPackageID: packageID,
		AmountPaid:    pkg.Price,
		PointsAwarded: pkg.RewardPoints,
		PurchaseDate:  time.Now(),
		Status:        "completed",
	}
	
	return s.CreditRepo.CreatePurchase(ctx, purchase)
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