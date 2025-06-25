package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
	"github.com/shawkyelshalawy/4sal-reward/internal/repositories"
)

type ProductService struct {
	ProductRepo *repositories.ProductRepository
}

func NewProductService(productRepo *repositories.ProductRepository) *ProductService {
	return &ProductService{
		ProductRepo: productRepo,
	}
}

func (s *ProductService) RedeemProduct(ctx context.Context, userID, productID uuid.UUID, quantity int) error {
	return s.ProductRepo.RedeemProduct(ctx, userID, productID, quantity)
}

func (s *ProductService) SearchProducts(ctx context.Context, query string, page, size int) ([]models.Product, error) {
	return s.ProductRepo.Search(ctx, query, page, size)
}

func (s *ProductService) CreateProduct(ctx context.Context, name, description string, categoryID *uuid.UUID, pointCost, stockQuantity int, isInOfferPool bool, imageURL string) (uuid.UUID, error) {
	product := &models.Product{
		ID:            uuid.New(),
		Name:          name,
		Description:   description,
		CategoryID:    categoryID,
		PointCost:     pointCost,
		StockQuantity: stockQuantity,
		IsActive:      true,
		IsInOfferPool: isInOfferPool,
		ImageURL:      imageURL,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err := s.ProductRepo.Create(ctx, product)
	return product.ID, err
}

func (s *ProductService) UpdateOfferStatus(ctx context.Context, productID uuid.UUID, isInOfferPool bool) error {
	return s.ProductRepo.UpdateOfferStatus(ctx, productID, isInOfferPool)
}