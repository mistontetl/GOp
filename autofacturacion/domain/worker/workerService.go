package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"portal_autofacturacion/data/queue"
	pac_service "portal_autofacturacion/domain/pac"
	ticket_service "portal_autofacturacion/domain/ticket"
	"portal_autofacturacion/models"
	"portal_autofacturacion/utils"

	"gorm.io/gorm"
)

const (
	MaxConcurrentJobs = 3
	MaxRetries        = 5
	TOLERANCIA        = 0.01
)

type invoiceWorker struct {
	Conexion         *gorm.DB
	sem              chan struct{}
	client           queue.RSGQueue
	TicketDataSource ticket_service.TicketDataSource
	PacDataSource    pac_service.PacDataSource
}

func NewInvoiceWorker(client queue.RSGQueue, ticketConfig string, pacConfig string, conexion *gorm.DB) invoiceWorker {
	return invoiceWorker{
		Conexion:         conexion,
		sem:              make(chan struct{}, MaxConcurrentJobs),
		client:           client,
		TicketDataSource: ticket_service.NewTicketDataSource(ticketConfig, conexion),
		PacDataSource:    pac_service.NewPacDataSource(pacConfig),
	}
}

func (w *invoiceWorker) HandleDelivery(delivery queue.Delivery) {
	log.Println("úruebassss;;;; Inicio HandleDelivery")
	w.sem <- struct{}{}
	defer func() { <-w.sem }()

	log.Println("[WORKER] Inicio HandleDelivery")

	var payload models.Payload
	if err := json.Unmarshal(delivery.Body(), &payload); err != nil {
		log.Println("Payload inválido:", err)
		delivery.Ack()
		return
	}

	log.Printf("Payload: UUID=%s Folio=%s Monto=%.2f Intento=%d",
		payload.UUID,
		payload.TicketFolio,
		payload.Total,
		payload.RetryCount,
	)

	//  Reclamar tarea
	fmt.Println(payload.UUID.String())

	res := w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ? AND status_id = ?", payload.UUID.String(), utils.PENDING).
		Updates(map[string]interface{}{
			"status_id":  utils.VALIDATING,
			"updated_at": time.Now(),
		})

	fmt.Println("respuesta ::: ", res.RowsAffected)

	if res.Error != nil || res.RowsAffected == 0 {
		log.Println("BillingRequest no reclamable o ya tomado")
		delivery.Ack()
		return
	}

	//  Buscar ticket en cache
	var ticketDB models.Ticket

	err := w.Conexion.
		Preload("Cliente").
		Preload("TicketLines").
		Where("id_ticket = ? AND is_sap = true", payload.TicketFolio).
		First(&ticketDB).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Error BD:", err)
		w.retry(payload, delivery)
		return
	}

	if err == nil {
		log.Println("Ticket encontrado en BD")

		if math.Abs(ticketDB.TotalAmount-payload.Total) > TOLERANCIA {
			log.Printf(" FRAUDE: DB=%.2f Usuario=%.2f", ticketDB.TotalAmount, payload.Total)
			errMsg := "Monto no coincide con el registro"
			w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			delivery.Ack()
			return
		}

		//	w.linkTicket(payload.UUID.String(), ticketDB.TkID)

		//	w.finishSuccess(payload.UUID.String())
		//	delivery.Ack()
		//	return
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("DB error leyendo ticket: %v", err)
		w.retry(payload, delivery)
		return
	}
	/* descomentar
	log.Println(" NO se encontró ticket en BD local, consultando CCO")
	// Fuente externa
	log.Println("Consultando fuente externa")
	extTicket, err := w.TicketDataSource.GetValidTicketForBilligS(payload.TicketFolio)
	*/
	if err != nil {
		if errors.Is(err, utils.ErrTicketNotFound) {
			errMsg := "Ticket no encontrado en fuente externa"
			w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			delivery.Ack()
			return
		}

		if utils.IsTemporaryError(err) || errors.Is(err, utils.ErrServerError) {
			w.retry(payload, delivery)
			return
		}

		errMsg := err.Error()
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	//  UPSERT
	/*descomentar
	extTicket.IdTicket = &payload.TicketFolio
	extTicket.IsSAP = true

	ticketDB, err = w.upsertTicket(extTicket)
	if err != nil {
		w.retry(payload, delivery)
		return
	}
	*/
	//  post-fetch
	if math.Abs(ticketDB.TotalAmount-payload.Total) > TOLERANCIA {
		log.Printf(" (EXTERNO): Externo=%.2f Usuario=%.2f", ticketDB.TotalAmount, payload.Total)

		errMsg := "Monto no coincide con el registro externo"
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	///
	//XML
	//validaciones de xml
	//  Validaciones OBLIGATORIAS antes de XML
	fmt.Println("validaciones de xml")
	if ticketDB.TkID == 0 {
		errMsg := "Ticket no encontrado"
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	if ticketDB.Cliente == nil {
		log.Println("Ticket sin cliente, asignando PUBLICO EN GENERAL")

		publicClient, err := w.genericClient()
		if err != nil {
			errMsg := "No se pudo obtener cliente generico"
			w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			delivery.Ack()
			return
		}

		// Asociar al ticket
		if err := w.Conexion.Model(&ticketDB).
			Update("cliente_id", publicClient.ClienteID).Error; err != nil {
			errMsg := "Error al asociar cliente público"
			w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			delivery.Ack()
			return
		}

		// Refrescar relaciones en memoria
		ticketDB.Cliente = publicClient
	}
	/*//verifica lineas
	if len(ticketDB.TicketLines) == 0 {
		errMsg := "Ticket sin conceptos"
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}
	*/
	fmt.Println("entra a xml")

	w.linkTicket(payload.UUID.String(), ticketDB.TkID)
	cfdiData := models.CFDIData{
		Ticket:  ticketDB,
		Cliente: *ticketDB.Cliente,
		Lineas:  ticketDB.TicketLines,
	}

	resp, err := w.PacDataSource.StampCFDI(cfdiData)
	if err != nil {
		errMsg := err.Error()
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}
	log.Println("UUID CFDI:", resp.UUID)
	log.Println("XML PATH:", resp.XMLPath)
	///
	w.updateStatus(payload.UUID.String(), utils.STAMPING, nil)

	time.Sleep(5 * time.Second)

	w.finishSuccess(payload.UUID.String())

	log.Printf("SUCCESS [%s] Ticket %s", payload.UUID, payload.TicketFolio)
	delivery.Ack()
}

func (w *invoiceWorker) StartConsuming(consumer queue.Consumer) error {
	log.Println(" StartConsuming llamado, esperando mensajes...")
	return consumer.ConsumeAsync(w.HandleDelivery)
}
func (w *invoiceWorker) retry(payload models.Payload, delivery queue.Delivery) {
	var br models.BillingRequest
	w.Conexion.Where("request_token = ?", payload.UUID.String()).First(&br)

	if br.BrRetryCount >= MaxRetries {
		errMsg := "Máximo de reintentos alcanzado"
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	newRetry := br.BrRetryCount + 1
	delay := int(math.Pow(2, float64(newRetry))) * 30

	w.updateStatusWithRetry(payload.UUID.String(), utils.PENDING, nil, newRetry)

	payload.RetryCount = newRetry
	w.client.PublishWithRetry(context.Background(), payload, delay)

	delivery.Ack()
}

func (w *invoiceWorker) upsertTicket(t models.Ticket) (models.Ticket, error) {
	var dbT models.Ticket

	err := w.Conexion.
		Where("id_ticket = ? AND is_sap = true", *t.IdTicket).
		First(&dbT).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return t, w.Conexion.Create(&t).Error
	}

	if err != nil {
		return models.Ticket{}, err
	}

	err = w.Conexion.Model(&dbT).
		Updates(map[string]interface{}{
			"total_amount": t.TotalAmount,
			"status":       t.Status,
			"updated_at":   time.Now(),
		}).Error

	return dbT, err
}

func (w *invoiceWorker) linkTicket(token string, tkID uint) {
	w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Update("ticket_id", tkID)
}

func (w *invoiceWorker) finishSuccess(token string) {
	w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Updates(map[string]interface{}{
			"status_id":  utils.SUCCESS,
			"updated_at": time.Now(),
		})
}

func (w *invoiceWorker) updateStatus(token string, status int, errMsg *string) {
	w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Updates(map[string]interface{}{
			"status_id":  status,
			"error":      errMsg,
			"updated_at": time.Now(),
		})
}

func (w *invoiceWorker) updateStatusWithRetry(token string, status int, errMsg *string, retryCount int) {
	w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Updates(map[string]interface{}{
			"status_id":      status,
			"error":          errMsg,
			"br_retry_count": retryCount,
			"updated_at":     time.Now(),
		})
}

/////Crea cliente

func (w *invoiceWorker) genericClient() (*models.Cliente, error) {
	var cliente models.Cliente

	err := w.Conexion.
		Where("rfc = ?", "XAXX010101000").
		First(&cliente).Error

	if err == nil {
		return &cliente, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	cliente = models.Cliente{
		Nombre:             "PUBLICO EN GENERAL",
		Email:              "facturacion@empresa.com",
		RFC:                "XAXX010101000",
		RegimenFiscal:      "616",
		DescripcionRegimen: "Sin obligaciones fiscales",
		PostalCode:         "00000",
		ExternalID:         "PUBLICO_GENERAL",
	}

	if err := w.Conexion.Create(&cliente).Error; err != nil {
		return nil, err
	}

	return &cliente, nil
}
