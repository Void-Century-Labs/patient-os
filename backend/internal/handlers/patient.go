package handlers

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/void-century-labs/patient-os/backend/internal/models"
	"gorm.io/gorm"
)

type PatientHandler struct {
	DB *gorm.DB
}

var mobileRegexp = regexp.MustCompile(`^\+?[1-9]\d{7,14}$`)

type registerPatientRequest struct {
	Name   string `json:"name" binding:"required"`
	Mobile string `json:"mobile" binding:"required"`
	Age    int    `json:"age"`
	Gender string `json:"gender"`
}

// Register creates a patient, or returns the existing patient for that
// mobile number so a returning patient can re-enter via QR without
// creating a duplicate identity.
func (h *PatientHandler) Register(c *gin.Context) {
	var req registerPatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !mobileRegexp.MatchString(req.Mobile) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mobile number"})
		return
	}

	var existing models.Patient
	if err := h.DB.Where("mobile = ?", req.Mobile).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, existing)
		return
	}

	patient := models.Patient{
		Name:   req.Name,
		Mobile: req.Mobile,
		Age:    req.Age,
		Gender: req.Gender,
	}
	if err := h.DB.Create(&patient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register patient"})
		return
	}

	c.JSON(http.StatusCreated, patient)
}

func (h *PatientHandler) Get(c *gin.Context) {
	var patient models.Patient
	if err := h.DB.First(&patient, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}
	c.JSON(http.StatusOK, patient)
}

func (h *PatientHandler) GetByMobile(c *gin.Context) {
	var patient models.Patient
	if err := h.DB.Where("mobile = ?", c.Query("mobile")).First(&patient).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}
	c.JSON(http.StatusOK, patient)
}

// GetNotifications lists a patient's notifications, most recent first, so
// the patient-facing app can surface queue status updates.
func (h *PatientHandler) GetNotifications(c *gin.Context) {
	var notifications []models.Notification
	if err := h.DB.Where("patient_id = ?", c.Param("id")).
		Order("created_at DESC").Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notifications"})
		return
	}
	c.JSON(http.StatusOK, notifications)
}
