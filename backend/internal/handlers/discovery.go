package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/void-century-labs/patient-os/backend/internal/models"
	"gorm.io/gorm"
)

type DiscoveryHandler struct {
	DB *gorm.DB
}

type doctorAvailability struct {
	models.Doctor
	QueueLength      int64 `json:"queue_length"`
	EstimatedWaitMin int64 `json:"estimated_wait_minutes"`
}

type departmentAvailability struct {
	models.Department
	Doctors []doctorAvailability `json:"doctors"`
}

// Departments lists a hospital's departments with each doctor's current
// queue length and estimated wait, for the patient-facing discovery screen.
func (h *DiscoveryHandler) Departments(c *gin.Context) {
	var departments []models.Department
	if err := h.DB.Where("hospital_id = ?", c.Param("hospitalID")).Find(&departments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch departments"})
		return
	}

	result := make([]departmentAvailability, 0, len(departments))
	for _, dept := range departments {
		var doctors []models.Doctor
		if err := h.DB.Where("department_id = ?", dept.ID).Find(&doctors).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch doctors"})
			return
		}

		availabilities := make([]doctorAvailability, 0, len(doctors))
		for _, doctor := range doctors {
			queueLength, err := h.waitingCount(doctor.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch queue length"})
				return
			}
			availabilities = append(availabilities, doctorAvailability{
				Doctor:           doctor,
				QueueLength:      queueLength,
				EstimatedWaitMin: queueLength * averageConsultMinutes,
			})
		}

		result = append(result, departmentAvailability{Department: dept, Doctors: availabilities})
	}

	c.JSON(http.StatusOK, result)
}

func (h *DiscoveryHandler) waitingCount(doctorID uint) (int64, error) {
	var count int64
	err := h.DB.Model(&models.QueueEntry{}).
		Joins("JOIN queues ON queues.id = queue_entries.queue_id").
		Where("queues.doctor_id = ? AND queue_entries.status = ?", doctorID, models.QueueEntryStatusWaiting).
		Count(&count).Error
	return count, err
}
