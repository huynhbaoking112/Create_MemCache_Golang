package models

import "gorm.io/gorm"

type Attendance struct {
	gorm.Model
	EmployeeID int
	Date       string `gorm:"type:date"`
	CheckIn    string `gorm:"type:time"`
	CheckOut   string `gorm:"type:time"`
	Shift      int
}
