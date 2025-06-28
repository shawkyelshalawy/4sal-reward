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
) {
	creditHandler := &handlers.CreditHandler{CreditService: creditService}
	productHandler := &handlers.ProductHandler{ProductService: productService}
	adminHandler := &handlers.AdminHandler{
		ProductService: productService,
		CreditService:  creditService,
	}	

	router.POST("/credits/purchase", creditHandler.PurchaseCreditPackage)
	router.POST("/products/redeem", productHandler.RedeemProduct)
	router.GET("/products/search", productHandler.SearchProducts)
	router.GET("/credits/packages", creditHandler.GetCreditPackages)

	admin := router.Group("/admin")
	{
		admin.POST("/packages", adminHandler.CreateCreditPackage)
		admin.POST("/products", adminHandler.CreateProduct)
		admin.PUT("/products/:id/offer-status", adminHandler.UpdateProductOfferStatus)
		admin.PUT("/packages/:id", adminHandler.UpdateCreditPackage)
		admin.PUT("/products/:id", adminHandler.UpdateProduct)
	}

}