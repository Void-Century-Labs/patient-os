package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/void-century-labs/patient-os/backend/internal/models"
	"github.com/void-century-labs/patient-os/backend/internal/ws"
	"gorm.io/gorm"
)

type QueueHandler struct {
	DB  *gorm.DB
	Hub *ws.Hub
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

	h.broadcastQueue(entry.QueueID)
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

	h.broadcastQueue(entry.QueueID)
	c.JSON(http.StatusOK, entry)
}

// CallNext marks the earliest waiting entry in a doctor's queue as called,
// for use by the queue operator dashboard.
func (h *QueueHandler) CallNext(c *gin.Context) {
	var queue models.Queue
	if err := h.DB.Where("doctor_id = ?", c.Param("id")).First(&queue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "queue not found"})
		return
	}

	var entry models.QueueEntry
	if err := h.DB.Where("queue_id = ? AND status = ?", queue.ID, models.QueueEntryStatusWaiting).
		Order("token_number").First(&entry).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no waiting patients"})
		return
	}

	entry.Status = models.QueueEntryStatusCalled
	if err := h.DB.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to call next patient"})
		return
	}

	h.notify(entry.PatientID, "queue_called", "You're up next — please head to the doctor's room.")
	h.broadcastQueue(entry.QueueID)
	c.JSON(http.StatusOK, entry)
}

// Complete marks a called entry as completed, freeing the doctor for the
// next patient.
func (h *QueueHandler) Complete(c *gin.Context) {
	var entry models.QueueEntry
	if err := h.DB.First(&entry, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "queue entry not found"})
		return
	}

	if entry.Status != models.QueueEntryStatusCalled {
		c.JSON(http.StatusConflict, gin.H{"error": "entry is not called"})
		return
	}

	entry.Status = models.QueueEntryStatusCompleted
	if err := h.DB.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete entry"})
		return
	}

	h.notify(entry.PatientID, "queue_completed", "Your visit is complete. Thank you!")
	h.broadcastQueue(entry.QueueID)
	c.JSON(http.StatusOK, entry)
}

// Skip marks a waiting or called entry as skipped.
func (h *QueueHandler) Skip(c *gin.Context) {
	var entry models.QueueEntry
	if err := h.DB.First(&entry, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "queue entry not found"})
		return
	}

	if entry.Status != models.QueueEntryStatusWaiting && entry.Status != models.QueueEntryStatusCalled {
		c.JSON(http.StatusConflict, gin.H{"error": "entry cannot be skipped"})
		return
	}

	entry.Status = models.QueueEntryStatusSkipped
	if err := h.DB.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to skip entry"})
		return
	}

	h.notify(entry.PatientID, "queue_skipped", "You were skipped. Please check in with the front desk.")
	h.broadcastQueue(entry.QueueID)
	c.JSON(http.StatusOK, entry)
}

type activeQueueEntry struct {
	models.QueueEntry
	PatientName string `json:"patient_name"`
	Position    int64  `json:"position"`
}

// ActiveQueue lists a doctor's currently called and waiting entries (with
// patient names and queue position) for the operator dashboard.
func (h *QueueHandler) ActiveQueue(c *gin.Context) {
	var doctor models.Doctor
	if err := h.DB.First(&doctor, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
		return
	}

	var queue models.Queue
	if err := h.DB.Where("doctor_id = ?", doctor.ID).First(&queue).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{"queue_id": nil, "entries": []activeQueueEntry{}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch queue"})
		return
	}

	var entries []models.QueueEntry
	if err := h.DB.Where(
		"queue_id = ? AND status IN ?",
		queue.ID, []models.QueueStatus{models.QueueEntryStatusWaiting, models.QueueEntryStatusCalled},
	).Order("status, token_number").Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch queue entries"})
		return
	}

	patientIDs := make([]uint, 0, len(entries))
	for _, entry := range entries {
		patientIDs = append(patientIDs, entry.PatientID)
	}
	var patients []models.Patient
	if len(patientIDs) > 0 {
		if err := h.DB.Where("id IN ?", patientIDs).Find(&patients).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch patients"})
			return
		}
	}
	nameByID := make(map[uint]string, len(patients))
	for _, patient := range patients {
		nameByID[patient.ID] = patient.Name
	}

	result := make([]activeQueueEntry, 0, len(entries))
	var position int64
	for _, entry := range entries {
		if entry.Status == models.QueueEntryStatusWaiting {
			position++
		}
		result = append(result, activeQueueEntry{
			QueueEntry:  entry,
			PatientName: nameByID[entry.PatientID],
			Position:    position,
		})
	}

	c.JSON(http.StatusOK, gin.H{"queue_id": queue.ID, "entries": result})
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

type queueUpdateMessage struct {
	Type    string           `json:"type"`
	QueueID uint             `json:"queue_id"`
	Entries []queueEntryView `json:"entries"`
}

// notify records a notification for a patient. Failures are logged but never
// block the queue state transition that triggered them.
func (h *QueueHandler) notify(patientID uint, notifType, message string) {
	h.DB.Create(&models.Notification{PatientID: patientID, Type: notifType, Message: message})
}

// broadcastQueue recalculates positions for all waiting entries in a queue
// and pushes the snapshot to every subscribed WebSocket client.
func (h *QueueHandler) broadcastQueue(queueID uint) {
	if h.Hub == nil {
		return
	}

	var entries []models.QueueEntry
	if err := h.DB.Where("queue_id = ? AND status = ?", queueID, models.QueueEntryStatusWaiting).
		Order("token_number").Find(&entries).Error; err != nil {
		return
	}

	views := make([]queueEntryView, 0, len(entries))
	for i, entry := range entries {
		views = append(views, queueEntryView{
			QueueEntry:       entry,
			Position:         int64(i + 1),
			EstimatedWaitMin: int64(i) * averageConsultMinutes,
		})
	}

	payload, err := json.Marshal(queueUpdateMessage{Type: "queue_update", QueueID: queueID, Entries: views})
	if err != nil {
		return
	}
	h.Hub.Broadcast(queueID, payload)
}
