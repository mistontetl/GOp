package migrations

import (
	"gorm.io/gorm"
)

func RunBillingMigrations(db *gorm.DB) {
	/*
	   	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

	   	// Crea tablas
	   	err := db.AutoMigrate(
	   		&models.BillingStatus{},
	   		&models.Billing_requests{},
	   		&models.BillingRequest{},
	   	)

	   	if err != nil {
	   		log.Fatal("Error al migrar billing tables:", err)
	   	}

	   	seedStatus(db)
	   }

	   func seedStatus(db *gorm.DB) {
	   	statuses := []models.BillingStatus{
	   		{Code: "PENDING", Description: "Peticion recibida y encolada"},
	   		{Code: "VALIDATING", Description: "Validando ticket"},
	   		{Code: "STAMPING", Description: "Enviando al PAC"},
	   		{Code: "SUCCESS", Description: "Factura generada"},
	   		{Code: "MAIL_FAILED", Description: "Error enviando correo"},
	   		{Code: "ERROR", Description: "Proceso fallido"},
	   	}

	   	for _, s := range statuses {
	   		db.FirstOrCreate(&s, models.BillingStatus{Code: s.Code})
	   	}
	*/
}
