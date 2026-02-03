package repository

import (
	"context"
	"portal_autofacturacion/models"
	"portal_autofacturacion/utils"

	"gorm.io/gorm"
)

/*
	TODO  -----------------  PERSISTENCE/ Open (Database, external APIs or files)  -----------------
*/

type BillingRequestRepository interface {
	Create(request *models.BillingRequest) error
	//
	FindByTicket(ticket string) (*models.BillingRequest, error)
	UpdateStatus(token string, statusID int, errMsg *string) error
	///
	ThereIsHistoryTicket(ctx context.Context, payload models.Payload) (models.Billing_requests, error)
	RegisterBillingHistory(ctx context.Context, payload models.Billing_requests) (models.Billing_requests, error)
}

type billingRequestRepository struct {
	db *gorm.DB
}

// //
func NewBillingRequestRepositoryPru(db *gorm.DB) BillingRequestRepository {
	return &billingRequestRepository{db: db}
}

// /
func (r *billingRequestRepository) FindByTicket(ticket string) (*models.BillingRequest, error) {
	var br models.BillingRequest

	err := r.db.
		Preload("Status").
		Where("user_input_ticket = ? and status_id != ?", ticket, utils.ERROR).
		Order("created_at DESC").
		First(&br).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &br, nil
}

func (r *billingRequestRepository) UpdateStatus(token string, statusID int, errMsg *string) error {
	return r.db.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Updates(map[string]interface{}{
			"status_id": statusID,
			"error":     errMsg,
		}).Error
}

// Inserta un registro en billing_requests
func (r *billingRequestRepository) Create(request *models.BillingRequest) error {
	return r.db.Create(request).Error
}

func (r *billingRequestRepository) UpdateStatus1(
	token string,
	statusID int,
	errMsg *string,
) error {

	return r.db.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Updates(map[string]interface{}{
			"status": statusID,
			"error":  errMsg,
		}).Error
}

////
/*
func NewBillingRequestRepository(db *sql.DB) BillingRequestRepository {
	return &billingRequestRepository{db: db}
}
*/
func (r billingRequestRepository) ThereIsHistoryTicket(ctx context.Context, payload models.Payload) (models.Billing_requests, error) {

	// There is a billing history for the ticket??? // USE -> ORM!!

	return models.Billing_requests{}, nil
}

func (r billingRequestRepository) RegisterBillingHistory(ctx context.Context, payload models.Billing_requests) (models.Billing_requests, error) {

	//"Create Ticket DB!!! //

	return models.Billing_requests{}, nil
}
