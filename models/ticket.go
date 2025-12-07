package models

import "time"

type Ticket struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:'abierto'"`
	Priority    string    `json:"priority" gorm:"default:'media'"`
	CreatedByID uint      `json:"createdById"`
	User        User      `json:"user" gorm:"foreignKey:CreatedByID"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
