package database

import (
	"fmt"
	"log"

	"github.com/sacolmenaress/apiTesisISUM/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB es una variable global para la conexión a la base de datos
var DB *gorm.DB

// ConnectDB inicializa la conexión a la base de datos
func ConnectDB() {
	var err error

	// Esta es la "cadena de conexión"
	// Reemplaza con tus datos de Docker o tu instalación local
	dsn := "host=127.0.0.1 user=myuser password=test1234 dbname=lc_db port=5432 sslmode=disable" // Conecta a la base de datos
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error al conectar a la base de datos: ", err)
	}

	fmt.Println("Conexión a la base de datos exitosa.")

	// --- Migración Automática ---
	// Esto crea las tablas en tu base de datos
	// basado en tus 'structs' en la carpeta /models
	// ¡Es la magia de GORM!
	err = DB.AutoMigrate(&models.User{}, &models.Ticket{}, &models.Incidencia{})
	if err != nil {
		log.Fatal("Error al migrar la base de datos: ", err)
	}

	fmt.Println("Migración de la base de datos exitosa.")
}
