package models

type BillingStatus struct {
	ID          uint   `gorm:"primaryKey"`
	Code        string `gorm:"size:20;unique;not null"`
	Description string
}
