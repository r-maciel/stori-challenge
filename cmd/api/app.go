package api

import (
	"database/sql"
	"log"
	"net/http"

	balanceapp "stori-challenge/internal/application/balance"
	csvmigration "stori-challenge/internal/application/csvmigration"
	infradb "stori-challenge/internal/infrastructure/db"
	"stori-challenge/internal/infrastructure/http/handlers"
	oas "stori-challenge/internal/infrastructure/http/openapi"

	"github.com/gin-gonic/gin"
)

// NewServer assembles and returns the HTTP server engine.
func NewServer() *gin.Engine {
	sqlDB, err := infradb.Open()
	if err != nil {
		log.Printf("database not available at startup: %v", err)
	}
	return NewServerWithDB(sqlDB)
}

// NewServerWithDB assembles the HTTP server engine using the provided DB connection.
func NewServerWithDB(sqlDB *sql.DB) *gin.Engine {
	router := gin.Default()

	// API versioning
	v1 := router.Group("/v1")

	// Health check
	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Infra wiring
	transactionRepo := infradb.NewTransactionRepo(sqlDB)
	migrationService := csvmigration.NewCsvMigrationService(transactionRepo)
	migrateHandler := handlers.NewMigrateHandler(migrationService)
	balanceService := balanceapp.NewBalanceService(transactionRepo)
	balanceHandler := handlers.NewBalanceHandler(balanceService)
	// Routes (v1)
	v1.POST("/migrate", migrateHandler.PostMigrate)
	v1.GET("/users/:user_id/balance", balanceHandler.GetBalance)

	// OpenAPI (3.1) documentation endpoints
	oas.RegisterOpenAPIRoutes(v1)

	return router
}
