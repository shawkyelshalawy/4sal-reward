package services

import (
	"context"

	"time"

	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
	"github.com/shawkyelshalawy/4sal-reward/internal/repositories"

	"github.com/google/uuid"
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