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