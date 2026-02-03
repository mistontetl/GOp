package models

import "time"

type Ticket struct {
	TkID               uint      `gorm:"tk_id;primaryKey;NOT NULL"`
	CancellationStatus string    `gorm:"cancellation_status"`
	ComentariosSAP     string    `gorm:"comentarios_sap"`
	CreateDate         time.Time `gorm:"column:create_date"`
	DateCreatedTicket  time.Time `gorm:"column:date_created_ticket"`
	ErrorSAP           string    `gorm:"column:error_sap"`
	ErrorTimbrado      string    `gorm:"column:error_timbrado"`
	FormaPago          string    `gorm:"column:forma_pago"`

	IdTicket      *string `gorm:"column:id_ticket"`
	IsGlobal      bool    `gorm:"column:is_global"`
	IsSAP         bool    `gorm:"column:is_sap"`
	ObjectKey     *string `gorm:"column:objectkey"`
	Status        string  `gorm:"column:status"`
	SystemGroupID string  `gorm:"column:system_group_id"`
	SystemID      string  `gorm:"column:system_id"`
	TotalAmount   float64 `gorm:"column:total_amount"`

	// FK
	ClienteID *uint    `gorm:"column:cliente_id"`
	Cliente   *Cliente `gorm:"foreignKey:ClienteID;references:ClienteID"`

	InsID   *uint    `gorm:"column:ins_id"`
	Invoice *Invoice `gorm:"foreignKey:InsID;references:InsID"`

	// N:M
	TicketLines []TicketLine `gorm:"foreignKey:TkID"`
	//
}

func (Ticket) TableName() string {
	return "tickets"
}
