package models

import (
	"time"

	"github.com/google/uuid"
)

type Payload struct {
	// Traking
	UUID uuid.UUID `json:"u"`
	//	BillingRequestID int64

	//Retries
	RetryCount int `json:"retry_count"`

	// Data
	TicketFolio string  `json:"f"`
	Total       float64 `json:"t"`
	RFC         string  `json:"r"`
	Email       string  `json:"e"`
}

type InvoiceTracking struct {
	UUID string `json:"u"`
}

type ResponseServerModel[T any] struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Res      T      `json:"res,omitempty"`
	Datetime string `json:"dateTime,omitempty"`
}

type TicketData struct {
	ID       string
	SourceID string // (Ej: "CCO", "GK-POST").

	IssueDate   time.Time
	GrossAmount float64
	Subtotal    float64
	TaxRate     float64
	TaxAmount   float64

	InvoiceUUID string
}

type TimbreResponse struct {
	UUID        string
	XMLTimbrado []byte
	XMLPath     string
}
