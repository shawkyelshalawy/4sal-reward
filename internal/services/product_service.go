package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
	"github.com/shawkyelshalawy/4sal-reward/internal/repositories"
)




type ProductService struct {
	ProductRepo *repositories.ProductRepository
	UserRepo    *repositories.UserRepository
}

func (s *ProductService) RedeemProduct(ctx context.Context, userID, productID uuid.UUID, quantity int) error {
	product, err := s.ProductRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	
	if !product.IsInOfferPool {
		return errors.New("product not available for redemption")
	}
	if product.StockQuantity < quantity {
		return errors.New("insufficient stock")
	}
	
	pointsNeeded := product.PointCost * quantity
	
	if err := s.UserRepo.DeductPoints(ctx, userID, pointsNeeded); err != nil {
		return err
	}
	
	if err := s.ProductRepo.ReduceStock(ctx, productID, quantity); err != nil {
		s.UserRepo.AddPoints(ctx, userID, pointsNeeded)
		return err
	}
	
	redemption := &models.PointRedemption{
		ID:             uuid.New(),
		UserID:         userID,
		ProductID:      productID,
		PointsUsed:     pointsNeeded,
		Quantity:       quantity,
		RedemptionDate: time.Now(),
		Status:         "completed",
	}
	
	return s.ProductRepo.CreateRedemption(ctx, redemption)
}