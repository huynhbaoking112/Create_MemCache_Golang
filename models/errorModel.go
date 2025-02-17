package models

import (
	"gorm.io/gorm"
)

type Error struct {
	gorm.Model
	EmployeeID int
	Date       string `gorm:"type:date"`
	Time       string `gorm:"type:time(3)"`
	TypeError  int
	IsPayment  string `gorm:"type:enum('OK','NO');default:'NO'"`
	Evidence   string `gorm:"type:text"`
}
