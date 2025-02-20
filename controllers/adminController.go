package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/global"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

}

func (*Admin) CreateBonus(c *gin.Context) {
	// lấy db
	db := global.Mdb

	// Móc body
	var body struct {
		EmployeeID  int
		Date        string
		Time        string
		Description string
		Money       float64
		IsPayment   string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to read body",
		})
		return
	}

	// Taọ Bonus
	result := db.Create(&models.Bonus{EmployeeID: body.EmployeeID, Date: body.Date, Time: body.Time, Description: body.Description, Money: body.Money, IsPayment: body.IsPayment})

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to create new of em error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Create new bonus of em success",
	})

}

func (*Admin) PaymentForEm(c *gin.Context) {
	// lấy db
	db := global.Mdb

	// {
	// 	EmployeeID : 1,
	// 	Date : "2025-02-18",
	// 	Time : "07:00:00",
	// 	Evidence : "url:asdasdsa.com",
	// 	attendance_id:[1,23,5,6],
	// 	bonus:[1,23,5,7],
	// 	error:[9,12,3]
	// }

	// định nghĩa struct để parse JSON từ request body
	var req struct {
		EmployeeID   int    `json:"EmployeeID"`
		Date         string `json:"Date"`
		Time         string `json:"Time"`
		Evidence     string `json:"Evidence"`
		AttendanceID []int  `json:"attendance_id"`
		Bonus        []int  `json:"bonus"`
		Error        []int  `json:"error"`
	}

	// Parse JSON từ request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Bắt đầu transaction để đảm bảo tính toàn vẹn dữ liệu
	tx := db.Begin()

	// 1️⃣ Tạo bản ghi mới trong bảng Payment
	payment := models.Payment{
		EmployeeID: req.EmployeeID,
		Date:       req.Date,
		Time:       req.Time,
		Evidence:   req.Evidence,
	}

	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		return
	}

	// Dùng goroutine
	var wg sync.WaitGroup
	errChan := make(chan error, 5)

	wg.Add(5)

	// 2️⃣ Cập nhật trạng thái is_payment trong bảng Bonus
	go func() {
		defer wg.Done()
		if len(req.Bonus) > 0 {
			if err := tx.Model(&models.Bonus{}).
				Where("id IN (?)", req.Bonus).
				Update("is_payment", "OK").Error; err != nil {
				errChan <- err
			}
		}
	}()

	// 3️⃣ Cập nhật trạng thái is_payment trong bảng Error
	go func() {
		defer wg.Done()
		if len(req.Error) > 0 {
			if err := tx.Model(&models.Error{}).
				Where("id IN (?)", req.Error).
				Update("is_payment", "OK").Error; err != nil {
				errChan <- err
			}
		}
	}()

	// 4️⃣ Tạo các bản ghi trong Payment_Infor

	go func() {
		defer wg.Done()
		var paymentInfoRecords []models.Payment_Infor
		for _, attendanceID := range req.AttendanceID {
			paymentInfoRecords = append(paymentInfoRecords, models.Payment_Infor{
				Id_payment:   int(payment.ID),
				AttendanceID: attendanceID,
			})
		}
		if len(paymentInfoRecords) > 0 {
			if err := tx.Create(&paymentInfoRecords).Error; err != nil {
				errChan <- err
			}
		}
	}()
	go func() {
		defer wg.Done()
		var paymentInfoRecords []models.Payment_Infor
		for _, bonusID := range req.Bonus {
			paymentInfoRecords = append(paymentInfoRecords, models.Payment_Infor{
				Id_payment: int(payment.ID),
				Bonus:      bonusID,
			})
		}
		if len(paymentInfoRecords) > 0 {
			if err := tx.Create(&paymentInfoRecords).Error; err != nil {
				errChan <- err
			}
		}
	}()
	go func() {
		defer wg.Done()
		var paymentInfoRecords []models.Payment_Infor
		for _, errorID := range req.Error {
			paymentInfoRecords = append(paymentInfoRecords, models.Payment_Infor{
				Id_payment: int(payment.ID),
				Error:      errorID,
			})
		}
		if len(paymentInfoRecords) > 0 {
			if err := tx.Create(&paymentInfoRecords).Error; err != nil {
				errChan <- err
			}
		}
	}()

	wg.Wait()
	close(errChan)

	for e := range errChan {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": e.Error()},
		)
		return
	}

	// Commit transaction nếu mọi thứ đều thành công
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":    "Payment processed successfully",
		"payment_id": payment.ID,
	})

}

func (*Admin) GetTakeLeave(c *gin.Context) {
	// Lấy DB
	db := global.Mdb

	// Lấy limit và offset từ query parameters (mặc định: limit=10, offset=2)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Khai báo slice để lưu kết quả
	var takeLeaves []models.TakeLeave

	// Truy vấn với limit và offset
	err := db.Limit(limit).Offset(offset).Find(&takeLeaves).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn cơ sở dữ liệu"})
		return
	}

	// Trả về kết quả
	c.JSON(http.StatusOK, gin.H{"data": takeLeaves})

}

func (*Admin) AcceptTakeleave(c *gin.Context) {
	// lấy db
	db := global.Mdb

	// Móc body
	var body struct {
		ID int
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to read body",
		})
		return
	}

	result := db.Model(&models.TakeLeave{}).Where("id = ?", body.ID).Update("is_agree", "OK")

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update take leave request",
		})
		return
	}

	// Phản hồi thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Take leave request approved successfully",
	})
}

func (*Admin) CreateSalary(c *gin.Context) {
	// Lấy DB
	db := global.Mdb

	// {
	// 	newSalary : 40000
	// }

	// Móc body
	var body struct {
		Salary float64 `json:"salary"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Lỗi móc dữ liệu",
		})
		return
	}

	// Taọ

	modelSalary := models.SalaryPartTime{Salary: body.Salary}
	if err := db.Create(&modelSalary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Lỗi tạo salary mới",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Taọ salary thành công",
	})

}

func (*Admin) ChangeSalary(c *gin.Context) {

	// lấy db
	db := global.Mdb

	// {
	// 	id_user : 1,
	// 	id_salary: 2
	// }

	// Móc body
	var body struct {
		Id_User   int `json:"id_user"`
		Id_Salary int `json:"id_salary"`
	}

	if err := c.ShouldBindBodyWithJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Dữ liệu đầu vào không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Kiểm tra id_salary có tồn tại không
	var salaryModel models.SalaryPartTime
	errS := db.Where("id = ?", body.Id_Salary).First(&salaryModel).Error

	if errS != nil {
		if errors.Is(errS, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Salary K TỔN TẠI",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Lỗi truy vấn cơ sở dữ liệu",
				"error":   errS.Error(),
			})
		}
		return
	}

	// Change salary cho user
	resultChange := db.Model(&models.Employee{}).Where("id = ?", body.Id_User).Update("salary_part_time", body.Id_Salary)

	if resultChange.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Cập nhật lương cho user thất bại",
			"error":   resultChange.Error.Error(),
		})
		return
	}

	// Kiểm tra nếu không có bản ghi nào bị ảnh hưởng
	if resultChange.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Không tìm thấy user để cập nhật",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (*Admin) ChangeActive(c *gin.Context) {
	db := global.Mdb

	// Móc dữ liệu

	// {
	// 	active:"YES"/"NONO"
	// }
	var body struct {
		Active string `json:"active"`
	}

	if err := c.ShouldBindBodyWithJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Dữ liệu đầu vào không đúng",
		})
		return
	}

	// lấy
	Idstr := c.Param("id")
	// Chuyển về dạng int
	id, err := strconv.Atoi(Idstr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Định dạng đường dẫn sai",
		})
		return
	}

	result := db.Model(&models.Employee{}).Where("id = ?", id).Update("is_active", body.Active)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Không thể cập nhật dữ liệu",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Id user không tồn tại",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Cập nhật hoạt động user thành công ",
	})

}
