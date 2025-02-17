package models

import "gorm.io/gorm"

type LimitEmployee struct {
	gorm.Model
	Date  string `gorm:"type:date"`
	Shift int    `gorm:"type:bigint"`
	Num   int    `gorm:"type:bigint"`
}
