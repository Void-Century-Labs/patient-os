package router

import (
	"github.com/gin-gonic/gin"
	"github.com/void-century-labs/patient-os/backend/internal/handlers"
	"gorm.io/gorm"
)

func New(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	health := &handlers.HealthHandler{DB: db}
	r.GET("/health", health.Check)

	r.Group("/api/v1")

	return r
}
