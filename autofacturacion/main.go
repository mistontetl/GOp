package main

import (
	"log"
	conexion "portal_autofacturacion/conexiones"
)

func main() {
	db, err := conexion.ConexionBD()

	if err != nil {
		log.Fatal(err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()
}
