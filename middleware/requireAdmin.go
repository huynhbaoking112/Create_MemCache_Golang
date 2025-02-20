package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/models"
)

func AdminCheck(c *gin.Context) {

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

	if userModel.Role != 2 {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Bạn không có quyền này",
		})
		return
	}

	c.Next()

}
