package models

import (
	"gorm.io/gorm"
)

type Payment_Infor struct {
	gorm.Model
	Id_payment   int `gorm:"type:bigint"`
	AttendanceID int `gorm:"type:bigint"`
	Bonus        int `gorm:"type:bigint"`
	Error        int `gorm:"type:bigint"`
}
