package conexion

import (
	"encoding/json"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type Config struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Port     string `json:"port"`
}

func ConectarSQLServer(rutaArchivo string) (*gorm.DB, error) {
	config, err := cargarConfig(rutaArchivo)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;encrypt=disable;TrustServerCertificate=true;dial timeout=30;",
		config.Host,
		config.User,
		config.Password,
		config.Port,
		config.Database,
	)

	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error al conectar a SQL Server: %v", err)
	}

	return db, nil
}

func ConexionBD() (*gorm.DB, error) {
	file, err := os.Open("config/postgre.json")
	if err != nil {
		return nil, fmt.Errorf("error al abrir config/postgre.json: %v", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("error al leer JSON: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Mexico_City",
		config.Host,
		config.User,
		config.Password,
		config.Database,
		config.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error al conectar a PostgreSQL: %v", err)
	}

	fmt.Println(" Conexión a PostgreSQL exitosa")
	return db, nil
}

func cargarConfig(ruta string) (Config, error) {
	var config Config
	file, err := os.Open(ruta)
	if err != nil {
		return config, fmt.Errorf("no se pudo abrir el archivo %s: %v", ruta, err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return config, fmt.Errorf("error al decodificar JSON %s: %v", ruta, err)
	}

	return config, nil
}
