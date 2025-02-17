package models

import "gorm.io/gorm"

type GroupEM struct {
	gorm.Model
	GroupID int
	EMID    int
}
