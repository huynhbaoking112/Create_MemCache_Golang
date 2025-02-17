package models

import "gorm.io/gorm"

type SalaryPartTime struct {
	gorm.Model
	Salary float64 `gorm:"type:decimal(10,2)"`
}
