package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/global"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/models"
)

type Common struct {
}

func GetCommon() *Common {
	return &Common{}
}

func (*Common) GetAttendance(c *gin.Context) {
	// lấy db:
	db := global.Mdb

	// Lấy thông tin user
	user, exits := c.Get("user")

	if !exits {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Bạn phải đăng nhập",
		})
		return
	}

	// Ép kiểu dữ liệu
	userModel, ok := user.(models.Employee)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to parse user data",
		})
		return
	}

	// Lấy giá trị id
	idStr := c.Param("id")
	// Ép kiểu id
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Kiểm tra quyền
	if userModel.Role != 2 && int(userModel.ID) != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền nàynày"})
		return
	}

	// Lấy attendance
	var userAttendace []models.Attendance

	result := db.Where("employee_id = ?", id).Find(&userAttendace)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Nếu không có dữ liệu, vẫn trả về một mảng rỗng
	c.JSON(http.StatusOK, gin.H{"attendances": userAttendace})
}
