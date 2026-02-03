package service

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"portal_autofacturacion/models"
	"time"
)

/*
type EdicomPac struct{}

func (EdicomPac) StampCFDI( //

	ticketData models.TicketData,
	//preStampedXML []byte, // XML

	) (models.TimbreResponse, error) {
		return models.TimbreResponse{}, fmt.Errorf("EDICOM!!!")
	}
*/
type EdicomPac struct{}

func (EdicomPac) StampCFDI(
	data models.CFDIData,
) (models.TimbreResponse, error) {

	log.Println("[EDICOM] StampCFDI - inicio")

	// Log entrada
	log.Printf("[EDICOM] Ticket ID: %d | Total: %.2f",
		data.Ticket.TkID,
		data.Ticket.TotalAmount,
	)

	log.Printf("[EDICOM] Cliente RFC: %s | Nombre: %s",
		data.Cliente.RFC,
		data.Cliente.Nombre,
	)

	log.Printf("[EDICOM] Líneas: %d", len(data.Lineas))

	//  Armar XML CFDI
	log.Println("[EDICOM] Paso 1 - BuildCFDI")
	xmlBytes, err := BuildCFDI(data)
	if err != nil {
		log.Println("[EDICOM][ERROR] BuildCFDI:", err)
		return models.TimbreResponse{}, fmt.Errorf("armado XML: %w", err)
	}

	log.Printf("[EDICOM] XML generado (%d bytes)", len(xmlBytes))

	//  UUID simulado
	uuid := fmt.Sprintf("UUID-TEST-%d", time.Now().Unix())
	log.Println("[EDICOM] UUID generado:", uuid)

	// 3Guardar XML
	path := fmt.Sprintf("cfdi_%s.xml", uuid)
	log.Println("[EDICOM] Guardando XML en:", path)

	if err := os.WriteFile(path, xmlBytes, 0644); err != nil {
		log.Println("[EDICOM][ERROR] Error al guardar XML:", err)
		return models.TimbreResponse{}, err
	}

	log.Println("[EDICOM] XML guardado correctamente")

	// Respuesta
	log.Println("[EDICOM] StampCFDI - fin OK")

	return models.TimbreResponse{
		UUID:        uuid,
		XMLTimbrado: xmlBytes,
		XMLPath:     path,
	}, nil
}

func BuildCFDI(data models.CFDIData) ([]byte, error) {
	log.Println("[BuildCFDI] inicio")

	ticket := data.Ticket
	cliente := data.Cliente
	lineas := data.Lineas

	log.Printf("[BuildCFDI] Ticket Total: %.2f | FormaPago: %s",
		ticket.TotalAmount,
		ticket.FormaPago,
	)

	log.Printf("[BuildCFDI] Cliente RFC: %s | Regimen: %s | CP: %s",
		cliente.RFC,
		cliente.RegimenFiscal,
		cliente.PostalCode,
	)

	if cliente.RFC == "" {
		log.Println("[BuildCFDI][ERROR] RFC vacio")
		return nil, fmt.Errorf("RFC receptor obligatorio")
	}
	/*
		if len(lineas) == 0 {
			log.Println("[BuildCFDI][ERROR] Sin lineas")
			return nil, fmt.Errorf("sin conceptos")
		}*/

	var conceptos []Concepto
	subTotal := 0.0

	log.Println("[BuildCFDI] Procesando conceptos")

	for i, l := range lineas {
		log.Printf(
			"[BuildCFDI] Línea %d | ProdServ=%s | Cant=%.2f | Base=%.2f",
			i+1,
			l.ClaveProdServ,
			l.Cantidad,
			l.Base,
		)

		subTotal += l.Base
		conceptos = append(conceptos, Concepto{
			ClaveProdServ:    l.ClaveProdServ,
			NoIdentificacion: l.NoIdentificacion,
			Cantidad:         fmt.Sprintf("%.2f", l.Cantidad),
			ClaveUnidad:      l.ClaveUnidad,
			Descripcion:      l.Descripcion,
			ValorUnitario:    fmt.Sprintf("%.2f", l.ValorUnitario),
			Importe:          fmt.Sprintf("%.2f", l.Base),
		})
	}

	log.Printf("[BuildCFDI] SubTotal calculado: %.2f", subTotal)

	cfdi := Comprobante{
		XmlnsCfdi:      "http://www.sat.gob.mx/cfd/4",
		XmlnsXsi:       "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: "http://www.sat.gob.mx/cfd/4 http://www.sat.gob.mx/sitio_internet/cfd/4/cfdv40.xsd",

		Version:           "4.0",
		Fecha:             time.Now().Format("2006-01-02T15:04:05"),
		FormaPago:         ticket.FormaPago,
		Moneda:            "MXN",
		SubTotal:          fmt.Sprintf("%.2f", subTotal),
		Total:             fmt.Sprintf("%.2f", ticket.TotalAmount),
		TipoDeComprobante: "I",
		MetodoPago:        "PUE",
		LugarExpedicion:   cliente.PostalCode,
		Exportacion:       "01",

		Emisor: Emisor{
			Rfc:           "AAA010101AAA",
			Nombre:        "EMPRESA PRUEBA SA DE CV",
			RegimenFiscal: "601",
		},

		Receptor: Receptor{
			Rfc:                     cliente.RFC,
			Nombre:                  cliente.Nombre,
			DomicilioFiscalReceptor: cliente.PostalCode,
			RegimenFiscalReceptor:   cliente.RegimenFiscal,
			UsoCFDI:                 "S01",
		},

		Conceptos: Conceptos{Concepto: conceptos},
	}

	log.Println("[BuildCFDI] Marshal XML")

	xmlBytes, err := xml.MarshalIndent(cfdi, "", "  ")
	if err != nil {
		log.Println("[BuildCFDI][ERROR] Marshal:", err)
		return nil, err
	}

	log.Println("[BuildCFDI] XML armado correctamente")

	return []byte(xml.Header + string(xmlBytes)), nil
}

type Comprobante struct {
	XMLName xml.Name `xml:"cfdi:Comprobante"`

	XmlnsCfdi      string `xml:"xmlns:cfdi,attr"`
	XmlnsXsi       string `xml:"xmlns:xsi,attr"`
	SchemaLocation string `xml:"xsi:schemaLocation,attr"`

	Version           string `xml:"Version,attr"`
	Fecha             string `xml:"Fecha,attr"`
	FormaPago         string `xml:"FormaPago,attr,omitempty"`
	SubTotal          string `xml:"SubTotal,attr"`
	Moneda            string `xml:"Moneda,attr"`
	Total             string `xml:"Total,attr"`
	TipoDeComprobante string `xml:"TipoDeComprobante,attr"`
	MetodoPago        string `xml:"MetodoPago,attr,omitempty"`
	LugarExpedicion   string `xml:"LugarExpedicion,attr"`
	Exportacion       string `xml:"Exportacion,attr"`

	Emisor    Emisor    `xml:"cfdi:Emisor"`
	Receptor  Receptor  `xml:"cfdi:Receptor"`
	Conceptos Conceptos `xml:"cfdi:Conceptos"`
}

type Emisor struct {
	Rfc           string `xml:"Rfc,attr"`
	Nombre        string `xml:"Nombre,attr"`
	RegimenFiscal string `xml:"RegimenFiscal,attr"`
}

type Receptor struct {
	Rfc                     string `xml:"Rfc,attr"`
	Nombre                  string `xml:"Nombre,attr"`
	DomicilioFiscalReceptor string `xml:"DomicilioFiscalReceptor,attr"`
	RegimenFiscalReceptor   string `xml:"RegimenFiscalReceptor,attr"`
	UsoCFDI                 string `xml:"UsoCFDI,attr"`
}

type Conceptos struct {
	Concepto []Concepto `xml:"cfdi:Concepto"`
}

type Concepto struct {
	ClaveProdServ    string `xml:"ClaveProdServ,attr"`
	NoIdentificacion string `xml:"NoIdentificacion,attr,omitempty"`
	Cantidad         string `xml:"Cantidad,attr"`
	ClaveUnidad      string `xml:"ClaveUnidad,attr"`
	Descripcion      string `xml:"Descripcion,attr"`
	ValorUnitario    string `xml:"ValorUnitario,attr"`
	Importe          string `xml:"Importe,attr"`
}
