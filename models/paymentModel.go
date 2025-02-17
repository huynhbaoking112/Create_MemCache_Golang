package models

import "gorm.io/gorm"

type Payment struct {
	gorm.Model
	EmployeeID   int
	AttendanceID int
	Date         string `gorm:"type:date"`
	Time         string `gorm:"type:time"`
	Evidence     string `gorm:"type:text"`
}
