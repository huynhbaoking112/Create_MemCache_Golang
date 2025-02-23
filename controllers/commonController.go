package controllers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/global"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/models"
	"golang.org/x/crypto/bcrypt"
)

type Common struct {
}

func GetCommon() *Common {
	return &Common{}
}

func (*Common) Login(c *gin.Context) {

	db := global.Mdb

	// Get the email and pass of req body
	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}

	// Look up requested user
	var user models.Employee
	db.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or  password",
		})

		return
	}

	// Compare sent in pass with saved user pass hash
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password",
		})

		return
	}

	// Generate a jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECERET")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid create token",
		})

		return
	}

	// send it back
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)

	// Send it back
	c.JSON(http.StatusOK, gin.H{
		"message": "Login success",
	})
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền này"})
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

func (cmm *Common) GetErrorOfEm(c *gin.Context) {
	db := global.Mdb

	// Xử lý id
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Xử lý user
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bạn phải đăng nhập"})
		return
	}
	userModel, ok := user.(models.Employee)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi ép kiểu dữ liệu"})
		return
	}

	// Kiểm tra quyền
	if userModel.ID != uint(id) && userModel.Role != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền này"})
		return
	}

	// Struct tạm thời để lưu kết quả
	type ErrorWithName struct {
		EmployeeID int
		Date       string
		Time       string
		TypeError  int
		IsPayment  string
		Evidence   string
		NameError  string
		Fines      float64
	}

	// Lấy lỗi
	var resultError []ErrorWithName
	result := db.Model(&models.Error{}).
		Select("errors.employee_id", "errors.date", "errors.time", "errors.type_error", "errors.is_payment", "errors.evidence", "error_names.name_error", "error_names.fines").
		Joins("left join error_names on error_names.id = errors.type_error").
		Where("errors.employee_id = ?", id).
		Order("errors.date DESC").
		Find(&resultError)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resultError})
}

func (cmm *Common) GetBonusOfEm(c *gin.Context) {
	db := global.Mdb

	// Xử lý id
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Xử lý user
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bạn phải đăng nhập"})
		return
	}
	userModel, ok := user.(models.Employee)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi ép kiểu dữ liệu"})
		return
	}

	// Kiểm tra quyền
	if userModel.ID != uint(id) && userModel.Role != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền này"})
		return
	}

	// Struct tạm thời để lưu kết quả
	type ErrorWithName struct {
		EmployeeID int
		Date       string
		Time       string
		TypeError  int
		IsPayment  string
		Evidence   string
		NameError  string
		Fines      float64
	}

	// Lấy lỗi
	var resultError []ErrorWithName
	result := db.Model(&models.Error{}).
		Select("errors.employee_id", "errors.date", "errors.time", "errors.type_error", "errors.is_payment", "errors.evidence", "error_names.name_error", "error_names.fines").
		Joins("left join error_names on error_names.id = errors.type_error").
		Where("errors.employee_id = ?", id).
		Order("errors.date DESC").
		Find(&resultError)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resultError})
}
