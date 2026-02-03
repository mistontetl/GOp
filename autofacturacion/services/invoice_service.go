package services

import (
	"errors"

	"portal_autofacturacion/models"

	"gorm.io/gorm"
)

type BillingService struct {
	db *gorm.DB
}

func (s *BillingService) CreateBillingRequest(req *models.BillingRequest) error {

	var status models.BillingStatus

	err := s.db.
		Where("code = ?", "PENDING").
		First(&status).Error

	if err != nil {
		return errors.New("estado PENDING no existe")
	}

	req.StatusID = status.ID
	return s.db.Create(req).Error
}
