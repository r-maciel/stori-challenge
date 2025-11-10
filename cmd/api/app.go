package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewServer assembles and returns the HTTP server engine.
func NewServer() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "oka",
		})
	})

	return router
}


