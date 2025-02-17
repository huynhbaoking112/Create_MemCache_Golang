package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	NameRole string `gorm:"type:varchar(50)"`
}
