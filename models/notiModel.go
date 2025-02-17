package models

import "gorm.io/gorm"

type Notification struct {
	gorm.Model
	EMID    int
	GroupID int
	DES     string `gorm:"type:text"`
}
