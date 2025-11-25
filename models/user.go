package models

import "time"

// User representa a un empleado de LC Consultores (Contador, Técnico, Admin)

type User struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	FirstName          string    `json:"firstName" gorm:"not null"`
	LastName           string    `json:"lastName" gorm:"not null"`
	Email              string    `json:"email" gorm:"unique;not null"`
	Password           string    `json:"-"`                    // El guion '-' evita que se envíe en respuestas JSON
	Role               string    `json:"role" gorm:"not null"` // Roles: "contador", "tecnico", "admin"
	MustChangePassword bool      `json:"mustChangePassword" gorm:"default:false"`
	CreatedAt          time.Time `json:"createdAt"`
}
