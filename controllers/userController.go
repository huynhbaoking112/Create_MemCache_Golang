package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/global"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/models"
)

type User struct {
}

func GetNewUser() *User {
	return &User{}
}

func (*User) RegisShift(c *gin.Context) {

	// Lấy DB
	db := global.Mdb

	var body struct {
		Shift int
		Date  string
	}

	// Móc dữ liệu
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}
	// Lấy giá trị user từ context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User not found in context",
		})
		return
	}

	// Ép kiểu user về models.Employee
	userModel, ok := user.(models.Employee)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to parse user data",
		})
		return
	}

	// Kiểm tra tổng đăng kí trong ngày và ca đó hiện tại
	var totalRegistrations int64

	// Truy vấn để đếm số lượng người đăng ký cho ngày cụ thể
	resultRegis := db.Model(&models.Registration{}).Where("date = ? AND shift = ?", body.Date, body.Shift).Count(&totalRegistrations)

	if resultRegis.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to count registrations",
		})
		return
	}

	// Kiểm tra có lớn hơn 6 không
	if totalRegistrations >= 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Employee in shift is full",
		})
		return
	}

	//kiểm tra limit employee
	var limitEm models.LimitEmployee

	db.Model(&models.LimitEmployee{}).Where("date = ? AND shift = ?", body.Date, body.Shift).First(&limitEm)

	// Nếu có limit
	if limitEm.Num != 0 {

		if limitEm.Num <= int(totalRegistrations) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Employee in shift is full",
			})
			return
		}

	}

	// Tạo đối tượng Registration
	registration := models.Registration{
		EmployeeID: int64(userModel.ID), // userModel.ID là kiểu int64
		Date:       body.Date,
		Shift:      body.Shift,
	}

	// Lưu vào DB
	result := db.Create(&registration)

	if result.Error != nil {
		fmt.Println(result.Error)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to register shift",
		})
		return
	}

	// Trả về response
	c.JSON(http.StatusOK, gin.H{
		"message": "Shift registered successfully",
	})
}
