package models

import "gorm.io/gorm"

type TakeLeave struct {
	gorm.Model
	EMID      int
	Date      string `gorm:"type:date"`
	Shift     int
	DES       string `gorm:"type:text"`
	Evidience string `gorm:"type:text"`
	IsAgree   string `gorm:"type:enum('OK','NO');default:'NO'"`
}
