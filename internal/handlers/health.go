package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func RegisterHealthRoutes(router *gin.Engine, db *sql.DB, redisClient *redis.Client) {
	router.GET("/health", func(c *gin.Context) {
		// Check database
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "DOWN",
				"db":     err.Error(),
			})
			return
		}

		ctx := c.Request.Context()
		if _, err := redisClient.Ping(ctx).Result(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "DOWN",
				"redis":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
			"db":     "OK",
			"redis":  "OK",
		})
	})
}