package models

import "gorm.io/gorm"

type ErrorName struct {
	gorm.Model
	NameError string  `gorm:"type:text"`
	Fines     float64 `gorm:"type:decimal(10,2)"`
}
