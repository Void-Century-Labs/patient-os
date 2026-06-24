package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/void-century-labs/patient-os/backend/internal/models"
	"gorm.io/gorm"
)

type HospitalHandler struct {
	DB *gorm.DB
}

type createHospitalRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *HospitalHandler) Create(c *gin.Context) {
	var req createHospitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hospital := models.Hospital{Name: req.Name}
	if err := h.DB.Create(&hospital).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create hospital"})
		return
	}

	c.JSON(http.StatusCreated, hospital)
}

func (h *HospitalHandler) List(c *gin.Context) {
	var hospitals []models.Hospital
	if err := h.DB.Find(&hospitals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch hospitals"})
		return
	}
	c.JSON(http.StatusOK, hospitals)
}

func (h *HospitalHandler) Get(c *gin.Context) {
	var hospital models.Hospital
	if err := h.DB.First(&hospital, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
		return
	}
	c.JSON(http.StatusOK, hospital)
}
