package main

import (
	//	"context"

	"context"
	"fmt"
	"log"
	"time"

	conexion "portal_autofacturacion/conexiones"
	"portal_autofacturacion/data/queue"
	"portal_autofacturacion/migrations"
	"portal_autofacturacion/models"

	"portal_autofacturacion/domain/worker"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	//WORKER!!!
	log.Println("✅ Worker started. Waiting for request...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	////

	db, err := conexion.ConexionBD()
	if err != nil {
		log.Fatalf("Error de conexión a la DB: %v", err)
	}
	///

	migrations.RunBillingMigrations(db)

	queueClient, err := queue.NewClient()

	if err != nil {
		log.Fatalf("RabbitMQ connection failed: %v", err)
	}
	defer queueClient.Close()

	tokenStr := "11015789-245b-4127-b456-14ff36fab85c"

	payloadPrueba := models.Payload{
		UUID:        uuid.MustParse(tokenStr),
		TicketFolio: "9001",
		Total:       1000.00,
	}

	fmt.Printf("Enviando mensaje para Ticket: %s...\n", payloadPrueba.TicketFolio)
	ctxPub, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := queueClient.Publish(ctxPub, payloadPrueba); err != nil {
		log.Printf(" Error al publicar mensaje de prueba: %v", err)
	} else {
		fmt.Println(" Mensaje de prueba enviado a la cola.")
	}
	if err != nil {
		log.Fatalf(" Error al publicar: %v", err)
	}

	fmt.Println(" Mensaje publicado con éxito.")

	///

	invoiceWorker := worker.NewInvoiceWorker(queueClient, "SAP", "PAC", db)
	//invoiceWorker := worker.NewInvoiceWorker(queueClient, "CCO", "PAC", db)

	if err := invoiceWorker.StartConsuming(queueClient); err != nil {
		log.Fatalf(" Error crítico al iniciar el Worker: %v", err)
	}

	log.Println(" Worker en ejecución. Esperando mensajes de RabbitMQ...")

	select {}
	///

}
