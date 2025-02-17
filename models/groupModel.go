package models

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Name string `gorm:"type:varchar(255)"`
}
