package models

import (
	"gorm.io/gorm"
)

type WorkShifts struct {
	gorm.Model
	ShiftName string `gorm:"type:varchar(50)"`
	StartTime string `gorm:"type:time(3)"` // Ép kiểu thành TIME(3)
	EndTime   string `gorm:"type:time(3)"` // Ép kiểu thành TIME(3)
}
