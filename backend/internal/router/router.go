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

	hospitals := &handlers.HospitalHandler{DB: db}
	departments := &handlers.DepartmentHandler{DB: db}
	doctors := &handlers.DoctorHandler{DB: db}
	patients := &handlers.PatientHandler{DB: db}
	discovery := &handlers.DiscoveryHandler{DB: db}

	api := r.Group("/api/v1")
	{
		api.POST("/hospitals", hospitals.Create)
		api.GET("/hospitals", hospitals.List)
		api.GET("/hospitals/:hospitalID", hospitals.Get)
		api.GET("/hospitals/:hospitalID/departments", departments.ListByHospital)
		api.POST("/hospitals/:hospitalID/departments", departments.Create)
		api.GET("/hospitals/:hospitalID/discovery", discovery.Departments)

		api.GET("/departments/:id", departments.Get)
		api.PATCH("/departments/:id", departments.Update)
		api.GET("/departments/:id/doctors", doctors.ListByDepartment)
		api.POST("/departments/:id/doctors", doctors.Create)

		api.GET("/doctors/:id", doctors.Get)
		api.PATCH("/doctors/:id/assign", doctors.Assign)

		api.POST("/patients/register", patients.Register)
		api.GET("/patients/:id", patients.Get)
		api.GET("/patients", patients.GetByMobile)
	}

	return r
}
