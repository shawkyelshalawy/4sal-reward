package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shawkyelshalawy/4sal-reward/internal/handlers"
	"github.com/shawkyelshalawy/4sal-reward/internal/infrastructure/cache"
	"github.com/shawkyelshalawy/4sal-reward/internal/infrastructure/db"
	"github.com/shawkyelshalawy/4sal-reward/internal/infrastructure/logger"
	"github.com/shawkyelshalawy/4sal-reward/internal/repositories"
	"github.com/shawkyelshalawy/4sal-reward/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	zapLog := logger.NewZapLogger()
	defer zapLog.Sync()

	pgDB, err := db.NewPostgresDB()
	if err != nil {
		zapLog.Fatal("Failed to connect to PostgreSQL", zapLog.ZapError(err))
	}
	defer pgDB.Close()
	
	redisClient := cache.NewRedisClient()
	defer redisClient.Close()

	if err := db.RunMigrations(); err != nil {
		zapLog.Fatal("Migrations failed", zapLog.ZapError(err))
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(zapLog.GinLogger())
	// Initialize repositories
	userRepo := repositories.NewUserRepository(pgDB)
	creditRepo := repositories.NewCreditRepository(pgDB)
	productRepo := repositories.NewProductRepository(pgDB)
	
	// Initialize services
	creditService := &services.CreditService{
		CreditRepo: creditRepo,
		UserRepo:   userRepo,
	}
	
	productService := &services.ProductService{
		ProductRepo: productRepo,
	}
	
	// Initialize handlers
	creditHandler := &handlers.CreditHandler{CreditService: creditService}
	productHandler := &handlers.ProductHandler{ProductService: productService}
	adminHandler := &handlers.AdminHandler{
		ProductService: productService,
		CreditService:  creditService,
	}
	aiHandler := &handlers.AIHandler{
		ProductService: productService,
		UserRepo:       userRepo,
	}
	
	// Public routes
	router.POST("/credits/purchase", creditHandler.PurchaseCreditPackage)
	router.POST("/products/redeem", productHandler.RedeemProduct)
	router.GET("/products/search", productHandler.SearchProducts)
	
	// Admin routes (in production, add authentication middleware)
	admin := router.Group("/admin")
	{
		admin.POST("/packages", adminHandler.CreateCreditPackage)
		admin.POST("/products", adminHandler.CreateProduct)
		admin.PUT("/products/:id/offer-status", adminHandler.UpdateProductOfferStatus)
	}
	
	// AI routes
	router.POST("/ai/recommendation", aiHandler.GetRecommendation)
	
	handlers.RegisterHealthRoutes(router, pgDB, redisClient)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLog.Fatal("Server failed", zapLog.ZapError(err))
		}
	}()

	zapLog.Info("Server started on :8080")
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLog.Info("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		zapLog.Fatal("Server forced to shutdown", zapLog.ZapError(err))
	}
	zapLog.Info("Server exited")
}