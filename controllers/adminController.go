package controllers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/global"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/models"
	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
}

func GetNewAdmin() *Admin {
	return &Admin{}
}
func (*Admin) Signup(c *gin.Context) {
	// Get the email/ pass off req body
	var body struct {
		FullName       string
		Gender         string
		Phone          string
		Address        string
		DateOfBirth    string
		HireDate       string
		Status         string
		Email          string
		Image          string
		CardNumber     string
		Bank           string
		IsActive       string
		Password       string
		Role           int
		SalaryPartTime int
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}
	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})

		return
	}

	// Create the user
	user := models.Employee{Email: body.Email, Password: string(hash), FullName: body.FullName, Gender: body.Gender, Phone: body.Phone, Address: body.Address, DateOfBirth: body.DateOfBirth, HireDate: body.HireDate, Image: body.Image, CardNumber: body.CardNumber, Bank: body.Bank, Role: body.Role, SalaryPartTime: body.SalaryPartTime}

	result := global.Mdb.Create(&user)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})

		return
	}

	// Respond
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})

}
func (*Admin) Login(c *gin.Context) {

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
		fmt.Println(err)
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
func (*Admin) Validate(c *gin.Context) {

	// Lấy giá trị user từ context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}

	// Ép kiểu user về models.User
	userModel, ok := user.(models.Employee)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse user data",
		})
		return
	}

	// In ra thông tin user
	// fmt.Println("User:", userModel)
	// fmt.Println("User:", userModel.)

	// Trả về response
	c.JSON(http.StatusOK, gin.H{
		"message": "Login Success",
		"user":    userModel,
	})
}
func (*Admin) LimitEm(c *gin.Context) {
	// Lấy db
	db := global.Mdb

	// Tạo struct body để nhận dữ liệu từ request body
	var body struct {
		Date  string `json:"date"` // Thêm json tag để chắc chắn rằng các trường sẽ được mapping đúng
		Shift int    `json:"shift"`
		Num   int    `json:"num"`
	}

	// Móc dữ liệu từ request body vào struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to read body",
			"details": err.Error(),
		})
		return
	}

	// Query tạo limit
	result := db.Create(&models.LimitEmployee{
		Date:  body.Date,
		Shift: body.Shift,
		Num:   body.Num,
	})

	// Kiểm tra lỗi khi tạo bản ghi
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to limit shift",
			"error":   result.Error.Error(),
		})
		return
	}

	// Trả về kết quả thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Limit shift successfully",
	})
}

func (*Admin) SetNewError(c *gin.Context) {

	// Set cơ sở dữ liệu
	db := global.Mdb

	// body
	var body struct {
		NameErr string
		Fines   float64
	}

	// Móc dữ liệu
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to read body",
		})
		return
	}

	// Tạo Lỗi
	result := db.Create(&models.ErrorName{NameError: body.NameErr, Fines: body.Fines})

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to create new error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Create new error success",
	})
	return

}

func (*Admin) HanldeErrorEm(c *gin.Context) {
	// Set cơ sở dữ liệu
	db := global.Mdb

	// body
	var body struct {
		Date       string
		EmployeeID int
		Time       string
		TypeError  int
		IsPayment  string
		Evidence   string
	}

	// Móc dữ liệu
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to read body",
		})
		return
	}

	// Tạo Lỗi
	result := db.Create(&models.Error{Date: body.Date, EmployeeID: body.EmployeeID, Time: body.Time, TypeError: body.TypeError, IsPayment: body.IsPayment, Evidence: body.Evidence})

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to create new of em error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Create new error of em success",
	})
	return
}
