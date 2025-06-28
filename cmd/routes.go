package main

import (
	"github.com/gin-gonic/gin"
	"github.com/shawkyelshalawy/4sal-reward/internal/handlers"
	"github.com/shawkyelshalawy/4sal-reward/internal/repositories"
	"github.com/shawkyelshalawy/4sal-reward/internal/services"
)

func SetupRoutes(
	router *gin.Engine,
	creditService *services.CreditService,
	productService *services.ProductService,
	userRepo *repositories.UserRepository,
	categoryRepo *repositories.CategoryRepository,
) {
	// Initialize handlers with proper dependency injection
	creditHandler := handlers.NewCreditHandler(creditService)
	productHandler := handlers.NewProductHandler(productService)
	adminHandler := &handlers.AdminHandler{
		ProductService: productService,
		CreditService:  creditService,
	}
	aiHandler := handlers.NewAIHandler(productService, userRepo, categoryRepo)

	// User endpoints
	router.POST("/credits/purchase", creditHandler.PurchaseCreditPackage)
	router.POST("/products/redeem", productHandler.RedeemProduct)
	router.GET("/products/search", productHandler.SearchProducts)
	router.GET("/credits/packages", creditHandler.GetCreditPackages)

	// AI recommendation endpoint
	router.POST("/ai/recommendation", aiHandler.GetRecommendation)

	// Admin endpoints
	admin := router.Group("/admin")
	{
		admin.POST("/packages", adminHandler.CreateCreditPackage)
		admin.POST("/products", adminHandler.CreateProduct)
		admin.PUT("/products/:id/offer-status", adminHandler.UpdateProductOfferStatus)
		admin.PUT("/packages/:id", adminHandler.UpdateCreditPackage)
		admin.PUT("/products/:id", adminHandler.UpdateProduct)
	}
}