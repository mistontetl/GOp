package models

import "time"

type Billing_requests struct {
	//Model!!!
	ID     int64
	Status int // Defualt PENDING (1) !!
}

type BillingRequest struct {
	RequestToken    string        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserInputTicket string        `gorm:"size:50;not null"`
	TotalAmount     float64       `gorm:"not null"`
	ClientEmail     string        `gorm:"size:150;not null"`
	RFC             string        `gorm:"size:13;not null"`
	StatusID        uint          `gorm:"not null"`
	Status          BillingStatus `gorm:"foreignKey:StatusID"`
	Error           *string       `gorm:"type:text"`
	BrRetryCount    int           `gorm:"column:br_retry_count"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
