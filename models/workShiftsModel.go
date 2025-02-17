package models

import (
	"time"

	"gorm.io/gorm"
)

type WorkShifts struct {
	gorm.Model
	ShiftName string    `gorm:"type:varchar(50)"`
	StartTime time.Time `gorm:"type:time(3)"` // Ép kiểu thành TIME(3)
	EndTime   time.Time `gorm:"type:time(3)"` // Ép kiểu thành TIME(3)
}
