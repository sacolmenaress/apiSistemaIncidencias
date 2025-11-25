// models/incidencia.go
package models

import "gorm.io/gorm"

// Incidencia representa una entrada de solución en la biblioteca
type Incidencia struct {
	gorm.Model        // Agrega campos: ID, CreatedAt, UpdatedAt, DeletedAt
	Title      string `gorm:"type:varchar(255);not null" json:"title"`
	Solution   string `gorm:"type:text;not null" json:"solution"`
	Category   string `gorm:"type:varchar(100);default:'General'" json:"category"`
	IsPublic   bool   `gorm:"default:true" json:"isPublic"`
}
