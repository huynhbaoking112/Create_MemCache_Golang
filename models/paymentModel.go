package models

import (
	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	EmployeeID int
	Date       string `gorm:"type:date"`
	Time       string `gorm:"type:time(3)"`
	Evidence   string `gorm:"type:text"`
}
