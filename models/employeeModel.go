package models

import "gorm.io/gorm"

type Employee struct {
	gorm.Model
	FullName       string `gorm:"type:varchar(150)"`
	Gender         string `gorm:"type:enum('Male','Female')"`
	Phone          string `gorm:"type:varchar(50)"`
	Address        string `gorm:"type:text"`
	DateOfBirth    string `gorm:"type:date"`
	HireDate       string `gorm:"type:date"`
	Email          string `gorm:"type:varchar(255);unique"`
	Image          string `gorm:"type:text"`
	CardNumber     string `gorm:"type:varchar(255)"`
	Bank           string `gorm:"type:varchar(255)"`
	NewNOTI        string `gorm:"type:enum('YES','NO');default:'NO'"`
	IsActive       string `gorm:"type:enum('YES','NO');default:'YES'"`
	Password       string `gorm:"type:varchar(255)"`
	Role           int    `gorm:"default:1"`
	SalaryPartTime int    `gorm:"default:0"`
}
