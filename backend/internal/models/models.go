package models

import "time"

type Hospital struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Departments []Department `json:"departments,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Department struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	HospitalID uint      `gorm:"not null;index" json:"hospital_id"`
	Name       string    `gorm:"not null" json:"name"`
	Doctors    []Doctor  `json:"doctors,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Doctor struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DepartmentID uint      `gorm:"not null;index" json:"department_id"`
	Name         string    `gorm:"not null" json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Patient struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Mobile    string    `gorm:"not null;uniqueIndex" json:"mobile"`
	Age       int       `json:"age"`
	Gender    string    `json:"gender"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type QueueStatus string

const (
	QueueEntryStatusWaiting   QueueStatus = "waiting"
	QueueEntryStatusCalled    QueueStatus = "called"
	QueueEntryStatusCompleted QueueStatus = "completed"
	QueueEntryStatusSkipped   QueueStatus = "skipped"
	QueueEntryStatusCancelled QueueStatus = "cancelled"
)

type Queue struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DoctorID  uint      `gorm:"not null;index" json:"doctor_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type QueueEntry struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	QueueID     uint        `gorm:"not null;index" json:"queue_id"`
	PatientID   uint        `gorm:"not null;index" json:"patient_id"`
	TokenNumber int         `gorm:"not null" json:"token_number"`
	Status      QueueStatus `gorm:"not null;default:waiting" json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PatientID uint      `gorm:"not null;index" json:"patient_id"`
	Type      string    `gorm:"not null" json:"type"`
	Message   string    `gorm:"not null" json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
