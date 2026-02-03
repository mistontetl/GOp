package models

import "time"

type TicketLine struct {
	TklID               uint       `gorm:"column:tkl_id;primaryKey;autoIncrement"`
	Amount              float64    `gorm:"column:amount"`
	Base                float64    `gorm:"column:base"`
	Cantidad            float64    `gorm:"column:cantidad"`
	ClaveProdServ       string     `gorm:"column:clave_prod_serv"`
	ClaveUnidad         string     `gorm:"column:clave_unidad"`
	DateCreate          *time.Time `gorm:"column:date_create"`
	Descripcion         string     `gorm:"column:descripcion"`
	Descuento           string     `gorm:"column:descuento"`
	NoIdentificacion    string     `gorm:"column:no_identificacion"`
	PorcentajeDescuento string     `gorm:"column:porcentaje_descuento"`
	TaxRate             float64    `gorm:"column:taxrate"`
	TaxRateTypeCode     string     `gorm:"column:tax_rate_type_code"`
	ValorUnitario       float64    `gorm:"column:valor_unitario"`

	//FK
	TkID uint `gorm:"column:tk_id"`
}

func (TicketLine) TableName() string {
	return "ticket_lines"
}
