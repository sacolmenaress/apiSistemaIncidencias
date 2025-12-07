package models

import "gorm.io/gorm"

type Incidencia struct {
	gorm.Model
	Title    string `gorm:"type:varchar(255);not null" json:"title"`
	Solution string `gorm:"type:text;not null" json:"solution"`
	Category string `gorm:"type:varchar(100);default:'General'" json:"category"`
	IsPublic bool   `gorm:"default:true" json:"isPublic"`
}
