package services

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/shawkyelshalawy/4sal-reward/internal/domain/models"
	"github.com/shawkyelshalawy/4sal-reward/internal/repositories"
)

const (
    productSearchCachePrefix = "product:search:"
    productSearchCacheTTL    = 5 * time.Minute
)

type ProductService struct {
    ProductRepo *repositories.ProductRepository
    RedisClient *redis.Client
}

func NewProductService(productRepo *repositories.ProductRepository, redisClient *redis.Client) *ProductService {
    return &ProductService{
        ProductRepo: productRepo,
        RedisClient: redisClient,
    }
}

func (s *ProductService) RedeemProduct(ctx context.Context, userID, productID uuid.UUID, quantity int) error {
    return s.ProductRepo.RedeemProduct(ctx, userID, productID, quantity)
}

func (s *ProductService) SearchProducts(ctx context.Context, query string, page, size int) ([]models.Product, error) {
    cacheKey := productSearchCachePrefix + query + ":" + strconv.Itoa(page) + ":" + strconv.Itoa(size)
    cachedData, err := s.RedisClient.Get(ctx, cacheKey).Result()
    if err == nil {
        var products []models.Product
        err = json.Unmarshal([]byte(cachedData), &products)
        if err == nil {
            return products, nil
        }
    } else if !errors.Is(err, redis.Nil) {
    }
    products, err := s.ProductRepo.Search(ctx, query, page, size)
    if err != nil {
        return nil, err
    }
    if len(products) > 0 {
        jsonData, err := json.Marshal(products)
        if err == nil {
            s.RedisClient.Set(ctx, cacheKey, jsonData, productSearchCacheTTL)
        }
    }
    return products, nil
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

func (s *ProductService) UpdateProduct(ctx context.Context, productID uuid.UUID, name, description *string, categoryID *uuid.UUID, pointCost, stockQuantity *int, isActive, isInOfferPool *bool, imageURL *string) error {
	s.clearProductCache(ctx)
	return s.ProductRepo.UpdateProduct(ctx, productID, name, description, categoryID, pointCost, stockQuantity, isActive, isInOfferPool, imageURL)
}

func (s *ProductService) GetProducts(ctx context.Context, page, size int, isActive, isInOfferPool, categoryID string) ([]models.Product, int, error) {
	return s.ProductRepo.GetProducts(ctx, page, size, isActive, isInOfferPool, categoryID)
}

// GetProductsByCategory retrieves products by category ID with pagination
func (s *ProductService) GetProductsByCategory(ctx context.Context, categoryID uuid.UUID, page, size int) ([]models.Product, int, error) {
	return s.ProductRepo.GetProductsByCategory(ctx, categoryID, page, size)
}

func (s *ProductService) clearProductCache(ctx context.Context) {
	// Get all keys matching the pattern
	keys, err := s.RedisClient.Keys(ctx, productSearchCachePrefix+"*").Result()
	if err != nil {
		return
	}
	
	if len(keys) > 0 {
		s.RedisClient.Del(ctx, keys...)
	}
}