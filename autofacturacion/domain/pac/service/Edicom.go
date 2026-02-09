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
	// Inicio de timbrado simulado: prepara XML y respuesta.
	log.Println("[EDICOM] StampCFDI - inicio")

	// Log de entrada con datos generales del ticket y cliente.
	log.Printf("[EDICOM] Ticket ID: %d | Total: %.2f",
		data.Ticket.TkID,
		data.Ticket.TotalAmount,
	)

	log.Printf("[EDICOM] Cliente RFC: %s | Nombre: %s",
		data.Cliente.RFC,
		data.Cliente.Nombre,
	)

	log.Printf("[EDICOM] Líneas: %d", len(data.Lineas))

	// Armar XML CFDI con los datos del ticket.
	log.Println("[EDICOM] Paso 1 - BuildCFDI")
	xmlBytes, err := BuildCFDI(data)
	if err != nil {
		log.Println("[EDICOM][ERROR] BuildCFDI:", err)
		return models.TimbreResponse{}, fmt.Errorf("armado XML: %w", err)
	}

	log.Printf("[EDICOM] XML generado (%d bytes)", len(xmlBytes))

	// UUID simulado para pruebas (sin PAC real).
	uuid := fmt.Sprintf("UUID-TEST-%d", time.Now().Unix())
	log.Println("[EDICOM] UUID generado:", uuid)

	// Guardar XML en disco para inspección local.
	path := fmt.Sprintf("cfdi_%s.xml", uuid)
	log.Println("[EDICOM] Guardando XML en:", path)

	if err := os.WriteFile(path, xmlBytes, 0644); err != nil {
		log.Println("[EDICOM][ERROR] Error al guardar XML:", err)
		return models.TimbreResponse{}, err
	}

	log.Println("[EDICOM] XML guardado correctamente")

	// Respuesta final con XML generado.
	log.Println("[EDICOM] StampCFDI - fin OK")

	return models.TimbreResponse{
		UUID:        uuid,
		XMLTimbrado: xmlBytes,
		XMLPath:     path,
	}, nil
}

func BuildCFDI(data models.CFDIData) ([]byte, error) {
	// Construcción del comprobante CFDI.
	log.Println("[BuildCFDI] inicio")

	// Extraer datos base del payload.
	ticket := data.Ticket
	cliente := data.Cliente
	lineas := data.Lineas

	// Validar datos mínimos del receptor.
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

	// Si no hay líneas, generar una línea de prueba para no fallar el XML.
	if len(lineas) == 0 {
		log.Println("[BuildCFDI][WARN] Sin lineas, generando datos de prueba")
		lineas = []models.TicketLine{
			{
				ClaveProdServ:    "90101501",
				NoIdentificacion: "CONSUMO",
				Cantidad:         1,
				ClaveUnidad:      "E48",
				Descripcion:      "CONSUMO DE ALIMENTOS",
				ValorUnitario:    129.310345,
				Base:             129.310345,
				TaxRate:          0.160000,
			},
		}
	}

	// Variables de acumulación para conceptos e impuestos.
	var conceptos []Concepto
	subTotal := 0.0
	totalImpuestosTrasladados := 0.0
	impuestosPorTasa := make(map[string]struct {
		base    float64
		importe float64
	})

	// Recorrer líneas para construir conceptos y sumar impuestos.
	log.Println("[BuildCFDI] Procesando conceptos")

	for i, l := range lineas {
		log.Printf(
			"[BuildCFDI] Línea %d | ProdServ=%s | Cant=%.2f | Base=%.2f",
			i+1,
			l.ClaveProdServ,
			l.Cantidad,
			l.Base,
		)

		// Sumar subtotal y calcular impuesto por línea.
		subTotal += l.Base
		lineaImpuesto := l.Base * l.TaxRate
		if lineaImpuesto > 0 {
			totalImpuestosTrasladados += lineaImpuesto
			tasaKey := fmt.Sprintf("%.2f", l.TaxRate)
			acumulado := impuestosPorTasa[tasaKey]
			acumulado.base += l.Base
			acumulado.importe += lineaImpuesto
			impuestosPorTasa[tasaKey] = acumulado
		}

		// Construir impuestos a nivel concepto si aplica.
		var impuestosConcepto *ImpuestosConcepto
		if lineaImpuesto > 0 {
			impuestosConcepto = &ImpuestosConcepto{
				Traslados: Traslados{
					Traslado: []Traslado{
						{
							Base:       fmt.Sprintf("%.2f", l.Base),
							Impuesto:   "002",
							TipoFactor: "Tasa",
							TasaOCuota: fmt.Sprintf("%.2f", l.TaxRate),
							Importe:    fmt.Sprintf("%.2f", lineaImpuesto),
						},
					},
				},
			}
		}

		// Agregar el concepto al comprobante.
		conceptos = append(conceptos, Concepto{
			ClaveProdServ:    l.ClaveProdServ,
			NoIdentificacion: l.NoIdentificacion,
			Cantidad:         fmt.Sprintf("%.2f", l.Cantidad),
			ClaveUnidad:      l.ClaveUnidad,
			Unidad:           "SERVICIO",
			Descripcion:      l.Descripcion,
			ValorUnitario:    fmt.Sprintf("%.2f", l.ValorUnitario),
			Importe:          fmt.Sprintf("%.2f", l.Base),
			ObjetoImp:        "02",
			Impuestos:        impuestosConcepto,
		})
	}

	log.Printf("[BuildCFDI] SubTotal calculado: %.2f", subTotal)

	// Construir el comprobante con emisor, receptor y conceptos.
	cfdi := Comprobante{
		XmlnsCfdi:      "http://www.sat.gob.mx/cfd/4",
		XmlnsXsi:       "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsTfd:       "http://www.sat.gob.mx/TimbreFiscalDigital",
		SchemaLocation: "http://www.sat.gob.mx/cfd/4 http://www.sat.gob.mx/sitio_internet/cfd/4/cfdv40.xsd",

		Version:           "4.0",
		Serie:             "DF3",
		Folio:             fmt.Sprintf("%d", ticket.TkID),
		Fecha:             time.Now().Format("2006-01-02T15:04:05"),
		Sello:             "",
		FormaPago:         ticket.FormaPago,
		NoCertificado:     "",
		Certificado:       "",
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

	// Construir impuestos globales agrupados por tasa.
	if totalImpuestosTrasladados > 0 {
		traslados := make([]Traslado, 0, len(impuestosPorTasa))
		for tasaKey, acumulado := range impuestosPorTasa {
			traslados = append(traslados, Traslado{
				Base:       fmt.Sprintf("%.2f", acumulado.base),
				Impuesto:   "002",
				TipoFactor: "Tasa",
				TasaOCuota: tasaKey,
				Importe:    fmt.Sprintf("%.2f", acumulado.importe),
			})
		}
		cfdi.Impuestos = &Impuestos{
			TotalImpuestosTrasladados: fmt.Sprintf("%.2f", totalImpuestosTrasladados),
			Traslados: Traslados{
				Traslado: traslados,
			},
		}
	}

	// Agregar complemento de timbre fiscal digital (placeholder).
	cfdi.Complemento = &Complemento{
		TimbreFiscalDigital: TimbreFiscalDigital{
			Version:          "1.1",
			RfcProvCertif:    "",
			UUID:             "",
			FechaTimbrado:    "",
			SelloCFD:         "",
			NoCertificadoSAT: "",
			SelloSAT:         "",
		},
	}

	// Serializar a XML con indentación.
	log.Println("[BuildCFDI] Marshal XML")

	xmlBytes, err := xml.MarshalIndent(cfdi, "", "  ")
	if err != nil {
		log.Println("[BuildCFDI][ERROR] Marshal:", err)
		return nil, err
	}

	// Retornar el XML con encabezado estándar.
	log.Println("[BuildCFDI] XML armado correctamente")

	return []byte(xml.Header + string(xmlBytes)), nil
}

type Comprobante struct {
	// Nodo raíz del CFDI.
	XMLName xml.Name `xml:"cfdi:Comprobante"`

	XmlnsCfdi      string `xml:"xmlns:cfdi,attr"`
	XmlnsXsi       string `xml:"xmlns:xsi,attr"`
	XmlnsTfd       string `xml:"xmlns:tfd,attr,omitempty"`
	SchemaLocation string `xml:"xsi:schemaLocation,attr"`

	Version           string `xml:"Version,attr"`
	Serie             string `xml:"Serie,attr,omitempty"`
	Folio             string `xml:"Folio,attr,omitempty"`
	Fecha             string `xml:"Fecha,attr"`
	Sello             string `xml:"Sello,attr,omitempty"`
	FormaPago         string `xml:"FormaPago,attr,omitempty"`
	NoCertificado     string `xml:"NoCertificado,attr,omitempty"`
	Certificado       string `xml:"Certificado,attr,omitempty"`
	SubTotal          string `xml:"SubTotal,attr"`
	Moneda            string `xml:"Moneda,attr"`
	Total             string `xml:"Total,attr"`
	TipoDeComprobante string `xml:"TipoDeComprobante,attr"`
	MetodoPago        string `xml:"MetodoPago,attr,omitempty"`
	LugarExpedicion   string `xml:"LugarExpedicion,attr"`
	Exportacion       string `xml:"Exportacion,attr"`

	Emisor      Emisor       `xml:"cfdi:Emisor"`
	Receptor    Receptor     `xml:"cfdi:Receptor"`
	Conceptos   Conceptos    `xml:"cfdi:Conceptos"`
	Impuestos   *Impuestos   `xml:"cfdi:Impuestos,omitempty"`
	Complemento *Complemento `xml:"cfdi:Complemento,omitempty"`
}

type Emisor struct {
	// Datos del emisor del comprobante.
	Rfc           string `xml:"Rfc,attr"`
	Nombre        string `xml:"Nombre,attr"`
	RegimenFiscal string `xml:"RegimenFiscal,attr"`
}

type Receptor struct {
	// Datos fiscales del receptor.
	Rfc                     string `xml:"Rfc,attr"`
	Nombre                  string `xml:"Nombre,attr"`
	DomicilioFiscalReceptor string `xml:"DomicilioFiscalReceptor,attr"`
	RegimenFiscalReceptor   string `xml:"RegimenFiscalReceptor,attr"`
	UsoCFDI                 string `xml:"UsoCFDI,attr"`
}

type Conceptos struct {
	// Lista de conceptos (partidas) del comprobante.
	Concepto []Concepto `xml:"cfdi:Concepto"`
}

type Concepto struct {
	// Información de cada partida del comprobante.
	ClaveProdServ    string             `xml:"ClaveProdServ,attr"`
	NoIdentificacion string             `xml:"NoIdentificacion,attr,omitempty"`
	Cantidad         string             `xml:"Cantidad,attr"`
	ClaveUnidad      string             `xml:"ClaveUnidad,attr"`
	Unidad           string             `xml:"Unidad,attr,omitempty"`
	Descripcion      string             `xml:"Descripcion,attr"`
	ValorUnitario    string             `xml:"ValorUnitario,attr"`
	Importe          string             `xml:"Importe,attr"`
	ObjetoImp        string             `xml:"ObjetoImp,attr,omitempty"`
	Impuestos        *ImpuestosConcepto `xml:"cfdi:Impuestos,omitempty"`
}

type ImpuestosConcepto struct {
	// Impuestos trasladados por concepto.
	Traslados Traslados `xml:"cfdi:Traslados"`
}

type Impuestos struct {
	// Impuestos globales del comprobante.
	TotalImpuestosTrasladados string    `xml:"TotalImpuestosTrasladados,attr,omitempty"`
	Traslados                 Traslados `xml:"cfdi:Traslados"`
}

type Traslados struct {
	// Lista de traslados (global o por concepto).
	Traslado []Traslado `xml:"cfdi:Traslado"`
}

type Traslado struct {
	// Datos del impuesto trasladado.
	Base       string `xml:"Base,attr"`
	Impuesto   string `xml:"Impuesto,attr"`
	TipoFactor string `xml:"TipoFactor,attr"`
	TasaOCuota string `xml:"TasaOCuota,attr"`
	Importe    string `xml:"Importe,attr"`
}

type Complemento struct {
	// Complemento del CFDI (timbre fiscal digital).
	TimbreFiscalDigital TimbreFiscalDigital `xml:"tfd:TimbreFiscalDigital"`
}

type TimbreFiscalDigital struct {
	// Datos del timbre fiscal digital (placeholder).
	Version          string `xml:"Version,attr"`
	RfcProvCertif    string `xml:"RfcProvCertif,attr,omitempty"`
	UUID             string `xml:"UUID,attr,omitempty"`
	FechaTimbrado    string `xml:"FechaTimbrado,attr,omitempty"`
	SelloCFD         string `xml:"SelloCFD,attr,omitempty"`
	NoCertificadoSAT string `xml:"NoCertificadoSAT,attr,omitempty"`
	SelloSAT         string `xml:"SelloSAT,attr,omitempty"`
}
