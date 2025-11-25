package models

import "time"

// Ticket representa una incidencia reportada en el sistema.
// Esto responde a los bloques A y B del cuestionario.
type Ticket struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`              // Título de la incidencia
	Description string    `json:"description" gorm:"not null"`        // Detalles del problema
	Status      string    `json:"status" gorm:"default:'abierto'"`    // Ej: abierto, en_proceso, resuelto
	Priority    string    `json:"priority" gorm:"default:'media'"`    // Ítem 2: "clasificación y priorización"
	CreatedByID uint      `json:"createdById"`                        // Quién lo creó
	User        User      `json:"user" gorm:"foreignKey:CreatedByID"` // Asociación al usuario
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
