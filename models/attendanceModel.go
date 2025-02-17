package models

import "gorm.io/gorm"

type Attendance struct {
	gorm.Model
	EmployeeID int
	Date       string `gorm:"type:date"`
	CheckIn    string `gorm:"type:time(3)"`
	CheckOut   string `gorm:"type:time(3)"`
	Shift      int
}
