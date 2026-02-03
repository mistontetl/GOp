package handler

import (
	"encoding/json"
	"fmt"
	"log"

	"net/http"
	"portal_autofacturacion/domain/api/controller"
	"portal_autofacturacion/models"
	"portal_autofacturacion/utils"
	"strings"
)

/*
	TODO  -----------------  HTTP -----------------
*/

/*
	TODO  -----------------  HTTP -----------------
*/
/////////////correcto

type BillingHistoryHandler struct {
	controller *controller.BillingRequestController
}

func NewTicketHandler(controller *controller.BillingRequestController) *BillingHistoryHandler {
	return &BillingHistoryHandler{controller: controller}
}

func (h *BillingHistoryHandler) TrackingBillingRequest(w http.ResponseWriter, r *http.Request) {
	//w.WriteHeader(http.StatusOK)

	var tracking models.InvoiceTracking

	if err := json.NewDecoder(r.Body).Decode(&tracking); err != nil {
		http.Error(w, "DTO NOT VALID", http.StatusBadRequest)
		return
	}

	utils.WriteAny(w, models.ResponseServerModel[any]{
		Code:     http.StatusOK,
		Datetime: utils.DateTime(),
		Res:      "OK",
	})
	//return
}

///

func (h *BillingHistoryHandler) CreateBillingRequest(
	w http.ResponseWriter,
	r *http.Request,
) {
	log.Println("[HTTP] CreateBillingRequest - request recibido")
	var req models.Payload

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		//http.Error(w, "Invalid DTO", http.StatusBadRequest)
		log.Printf("[HTTP][ERROR] DTO inválido err=%v", err)
		utils.WriteErr(w, "", fmt.Errorf("Invalid DTO"), http.StatusConflict)
		return
	}
	fmt.Println("CAMPOS ::::::::::::::: ::::::::: ", req)
	//  SOLO crea la factura

	token, isNew, key, err := h.controller.CreateInvoice(req)
	if err != nil {
		if strings.Contains(err.Error(), "409 Conflict") {
			utils.WriteErr(w, "", err, http.StatusConflict)
			//	http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		utils.WriteErr(w, "", err, http.StatusBadRequest)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(key)
	//	status := http.StatusOK

	if isNew {
		req.UUID = *key
		req.RetryCount = 0
		fmt.Println("::  ", req)
		if err := h.controller.PublishToQueue(r.Context(), req); err != nil {
			utils.WriteErr(w, "", err, http.StatusConflict)
			return
		}
		//	status = http.StatusAccepted // 202
	}

	status := http.StatusOK
	if isNew {
		status = http.StatusAccepted // 202
	}

	log.Printf(
		"[invoice] token=%s ticket=%s new=%v",
		token, req.TicketFolio, isNew,
	)
	///	req.UUID = *key
	//	req.RetryCount = 0
	/*
		err = h.controller.PublishToQueue(r.Context(), req)
		if err != nil {
			utils.WriteErr(w, "", err, http.StatusConflict)
			return
			//	http.Error(w, err.Error(), http.StatusConflict)
		}*/
	utils.WriteAny(w, models.ResponseServerModel[any]{
		Code:     status,
		Datetime: utils.DateTime(),
		Res: map[string]string{
			"request_token": token.RequestToken,
		},
	})
}
