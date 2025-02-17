package models

import "gorm.io/gorm"

type Bonus struct {
	gorm.Model
	EmployeeID  int
	Date        string  `gorm:"type:date"`
	Time        string  `gorm:"type:time"`
	Description float64 `gorm:"type:decimal(10,2)"`
	Money       float64 `gorm:"type:decimal(10,2)"`
	IsPayment   string  `gorm:"type:enum('OK','NO')"`
}
