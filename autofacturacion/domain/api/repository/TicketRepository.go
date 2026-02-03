package repository

import (
	"portal_autofacturacion/domain/ticket/dto"
	"portal_autofacturacion/models"
	"strconv"
	"time"

	"gorm.io/gorm"
)

func SaveTicketServer(db *gorm.DB, rows []dto.TicketRow) (*models.Ticket, error) {

	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var ticketCreado models.Ticket

	err := db.Transaction(func(tx *gorm.DB) error {
		header := rows[0]

		ticket := models.Ticket{
			IdTicket:      &header.IDTicket,
			Status:        header.Status,
			FormaPago:     header.FormaPago,
			SystemID:      header.SystemID,
			SystemGroupID: header.SystemGroupID,
			TotalAmount:   header.TotalAmount,
			ObjectKey:     &header.ObjectKey,

			CancellationStatus: strconv.Itoa(header.CancellationStatus),
			DateCreatedTicket:  header.DateCreated,
			CreateDate:         time.Now(),
		}

		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}

		for _, r := range rows {
			now := time.Now()

			line := models.TicketLine{
				TkID:                ticket.TkID,
				Amount:              r.Amount,
				Base:                r.Base,
				Cantidad:            r.Cantidad,
				ClaveProdServ:       r.ClaveProdServ,
				ClaveUnidad:         r.ClaveUnidad,
				DateCreate:          &now,
				Descripcion:         r.Descripcion,
				Descuento:           r.Descuento,
				NoIdentificacion:    r.NoIdentificacion,
				PorcentajeDescuento: r.PorcentajeDescuento,
				TaxRate:             r.Taxrate,
				TaxRateTypeCode:     r.TaxrateTypeCode,
				ValorUnitario:       r.ValorUnitario,
			}

			if err := tx.Create(&line).Error; err != nil {
				return err
			}

		}
		ticketCreado = ticket
		return nil

	})

	if err != nil {
		return nil, err
	}

	return &ticketCreado, nil
}
