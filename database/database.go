package database

import (
	"fmt"
	"log"

	"github.com/sacolmenaress/apiTesisISUM/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	dsn := "host=127.0.0.1 user=myuser password=test1234 dbname=lc_db port=5432 sslmode=disable" // Conecta a la base de datos
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error al conectar a la base de datos: ", err)
	}

	fmt.Println("Conexión a la base de datos exitosa.")

	err = DB.AutoMigrate(&models.User{}, &models.Ticket{}, &models.Incidencia{})
	if err != nil {
		log.Fatal("Error al migrar la base de datos: ", err)
	}

	fmt.Println("Migración de la base de datos exitosa.")
}
