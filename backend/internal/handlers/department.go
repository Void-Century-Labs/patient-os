package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/void-century-labs/patient-os/backend/internal/models"
	"gorm.io/gorm"
)

type DepartmentHandler struct {
	DB *gorm.DB
}

type createDepartmentRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *DepartmentHandler) Create(c *gin.Context) {
	var req createDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var hospital models.Hospital
	if err := h.DB.First(&hospital, c.Param("hospitalID")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
		return
	}

	dept := models.Department{HospitalID: hospital.ID, Name: req.Name}
	if err := h.DB.Create(&dept).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create department"})
		return
	}

	c.JSON(http.StatusCreated, dept)
}

func (h *DepartmentHandler) ListByHospital(c *gin.Context) {
	var departments []models.Department
	if err := h.DB.Where("hospital_id = ?", c.Param("hospitalID")).Find(&departments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch departments"})
		return
	}
	c.JSON(http.StatusOK, departments)
}

func (h *DepartmentHandler) Get(c *gin.Context) {
	var dept models.Department
	if err := h.DB.First(&dept, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
		return
	}
	c.JSON(http.StatusOK, dept)
}

type updateDepartmentRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *DepartmentHandler) Update(c *gin.Context) {
	var dept models.Department
	if err := h.DB.First(&dept, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
		return
	}

	var req updateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dept.Name = req.Name
	if err := h.DB.Save(&dept).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update department"})
		return
	}

	c.JSON(http.StatusOK, dept)
}
