// Command seed populates a clean demo hospital with departments and doctors
// so the app has something to show immediately after a fresh deploy. It is
// idempotent: re-running it will not create duplicates.
package main

import (
	"log"

	"github.com/void-century-labs/patient-os/backend/internal/config"
	"github.com/void-century-labs/patient-os/backend/internal/database"
	"github.com/void-century-labs/patient-os/backend/internal/models"
	"gorm.io/gorm"
)

type departmentSeed struct {
	Name    string
	Doctors []string
}

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	hospital, err := findOrCreateHospital(db, "Demo General Hospital")
	if err != nil {
		log.Fatalf("failed to seed hospital: %v", err)
	}

	departments := []departmentSeed{
		{Name: "Cardiology", Doctors: []string{"Dr. Patel", "Dr. Nguyen"}},
		{Name: "Pediatrics", Doctors: []string{"Dr. Okafor"}},
		{Name: "Orthopedics", Doctors: []string{"Dr. Romero"}},
	}

	for _, dept := range departments {
		department, err := findOrCreateDepartment(db, hospital.ID, dept.Name)
		if err != nil {
			log.Fatalf("failed to seed department %s: %v", dept.Name, err)
		}
		for _, doctorName := range dept.Doctors {
			if err := findOrCreateDoctor(db, department.ID, doctorName); err != nil {
				log.Fatalf("failed to seed doctor %s: %v", doctorName, err)
			}
		}
	}

	log.Printf("Seeded hospital %q (id=%d) with %d departments", hospital.Name, hospital.ID, len(departments))
}

func findOrCreateHospital(db *gorm.DB, name string) (models.Hospital, error) {
	var hospital models.Hospital
	err := db.Where("name = ?", name).First(&hospital).Error
	if err == nil {
		return hospital, nil
	}
	if err != gorm.ErrRecordNotFound {
		return hospital, err
	}
	hospital = models.Hospital{Name: name}
	return hospital, db.Create(&hospital).Error
}

func findOrCreateDepartment(db *gorm.DB, hospitalID uint, name string) (models.Department, error) {
	var department models.Department
	err := db.Where("hospital_id = ? AND name = ?", hospitalID, name).First(&department).Error
	if err == nil {
		return department, nil
	}
	if err != gorm.ErrRecordNotFound {
		return department, err
	}
	department = models.Department{HospitalID: hospitalID, Name: name}
	return department, db.Create(&department).Error
}

func findOrCreateDoctor(db *gorm.DB, departmentID uint, name string) error {
	var doctor models.Doctor
	err := db.Where("department_id = ? AND name = ?", departmentID, name).First(&doctor).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	doctor = models.Doctor{DepartmentID: departmentID, Name: name}
	return db.Create(&doctor).Error
}
