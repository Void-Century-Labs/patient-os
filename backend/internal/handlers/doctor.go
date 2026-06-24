package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/void-century-labs/patient-os/backend/internal/models"
	"gorm.io/gorm"
)

type DoctorHandler struct {
	DB *gorm.DB
}

type createDoctorRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *DoctorHandler) Create(c *gin.Context) {
	var dept models.Department
	if err := h.DB.First(&dept, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
		return
	}

	var req createDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doctor := models.Doctor{DepartmentID: dept.ID, Name: req.Name}
	if err := h.DB.Create(&doctor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create doctor"})
		return
	}

	c.JSON(http.StatusCreated, doctor)
}

func (h *DoctorHandler) ListByDepartment(c *gin.Context) {
	var doctors []models.Doctor
	if err := h.DB.Where("department_id = ?", c.Param("id")).Find(&doctors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch doctors"})
		return
	}
	c.JSON(http.StatusOK, doctors)
}

func (h *DoctorHandler) Get(c *gin.Context) {
	var doctor models.Doctor
	if err := h.DB.First(&doctor, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
		return
	}
	c.JSON(http.StatusOK, doctor)
}

type assignDoctorRequest struct {
	DepartmentID uint `json:"department_id" binding:"required"`
}

func (h *DoctorHandler) Assign(c *gin.Context) {
	var doctor models.Doctor
	if err := h.DB.First(&doctor, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
		return
	}

	var req assignDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var dept models.Department
	if err := h.DB.First(&dept, req.DepartmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
		return
	}

	doctor.DepartmentID = dept.ID
	if err := h.DB.Save(&doctor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign doctor"})
		return
	}

	c.JSON(http.StatusOK, doctor)
}
