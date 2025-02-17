package models

import "gorm.io/gorm"

// Mô hình Registration
type Registration struct {
	gorm.Model
	EmployeeID int64  `gorm:"column:employee_id;type:bigint"` // Đảm bảo kiểu dữ liệu là int64
	Date       string `gorm:"type:date"`
	Shift      int    `gorm:"type:bigint"`
}
