package models

import "gorm.io/gorm"

type Error struct {
	gorm.Model
	EmployeeID int
	Date       string `gorm:"type:date"`
	Time       string `gorm:"type:time"`
	TypeError  int
	IsPayment  string `gorm:"type:enum('OK','NO')"`
	Evidence   string `gorm:"type:text"`
}
