package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/void-century-labs/patient-os/backend/internal/models"
	"gorm.io/gorm"
)

type QueueHandler struct {
	DB *gorm.DB
}

// averageConsultMinutes is a fixed placeholder used to estimate wait time
// until real consultation duration data is collected.
const averageConsultMinutes = 10

type queueEntryView struct {
	models.QueueEntry
	Position         int64 `json:"position"`
	EstimatedWaitMin int64 `json:"estimated_wait_minutes"`
}

type joinQueueRequest struct {
	PatientID uint `json:"patient_id" binding:"required"`
}

// Join finds or creates the queue for a doctor, generates the next token
// number, and adds the patient as a waiting entry. A patient can only
// hold one active (waiting/called) entry per queue at a time.
func (h *QueueHandler) Join(c *gin.Context) {
	var doctor models.Doctor
	if err := h.DB.First(&doctor, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
		return
	}

	var req joinQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var patient models.Patient
	if err := h.DB.First(&patient, req.PatientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}

	var entry models.QueueEntry
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var queue models.Queue
		if err := tx.Where("doctor_id = ?", doctor.ID).First(&queue).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			queue = models.Queue{DoctorID: doctor.ID}
			if err := tx.Create(&queue).Error; err != nil {
				return err
			}
		}

		var existing models.QueueEntry
		err := tx.Where(
			"queue_id = ? AND patient_id = ? AND status IN ?",
			queue.ID, patient.ID, []models.QueueStatus{models.QueueEntryStatusWaiting, models.QueueEntryStatusCalled},
		).First(&existing).Error
		if err == nil {
			entry = existing
			return nil
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		var lastToken int
		if err := tx.Model(&models.QueueEntry{}).
			Where("queue_id = ?", queue.ID).
			Select("COALESCE(MAX(token_number), 0)").
			Scan(&lastToken).Error; err != nil {
			return err
		}

		entry = models.QueueEntry{
			QueueID:     queue.ID,
			PatientID:   patient.ID,
			TokenNumber: lastToken + 1,
			Status:      models.QueueEntryStatusWaiting,
		}
		return tx.Create(&entry).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join queue"})
		return
	}

	view, err := h.toView(entry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute queue status"})
		return
	}

	c.JSON(http.StatusCreated, view)
}

// Leave cancels a waiting entry. Entries already called or completed
// cannot be left.
func (h *QueueHandler) Leave(c *gin.Context) {
	var entry models.QueueEntry
	if err := h.DB.First(&entry, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "queue entry not found"})
		return
	}

	if entry.Status != models.QueueEntryStatusWaiting {
		c.JSON(http.StatusConflict, gin.H{"error": "entry is not waiting"})
		return
	}

	entry.Status = models.QueueEntryStatusCancelled
	if err := h.DB.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to leave queue"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// Status returns the entry's current position among waiting entries and
// an estimated wait time, for the patient-facing tracking screen.
func (h *QueueHandler) Status(c *gin.Context) {
	var entry models.QueueEntry
	if err := h.DB.First(&entry, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "queue entry not found"})
		return
	}

	view, err := h.toView(entry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute queue status"})
		return
	}

	c.JSON(http.StatusOK, view)
}

func (h *QueueHandler) toView(entry models.QueueEntry) (queueEntryView, error) {
	if entry.Status != models.QueueEntryStatusWaiting {
		return queueEntryView{QueueEntry: entry, Position: 0, EstimatedWaitMin: 0}, nil
	}

	var position int64
	err := h.DB.Model(&models.QueueEntry{}).
		Where("queue_id = ? AND status = ? AND token_number <= ?", entry.QueueID, models.QueueEntryStatusWaiting, entry.TokenNumber).
		Count(&position).Error
	if err != nil {
		return queueEntryView{}, err
	}

	return queueEntryView{
		QueueEntry:       entry,
		Position:         position,
		EstimatedWaitMin: (position - 1) * averageConsultMinutes,
	}, nil
}
