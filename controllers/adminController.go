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

// Hằng số mặc định cho giới hạn số lượng nhân viên mỗi ca
const DefaultShiftLimit = 6

// LimitEm giới hạn số lượng nhân viên trong một ca làm việc
func (*Admin) LimitEm(c *gin.Context) {
	// Get the db
	db := global.Mdb

	// Parse request body
	var req struct {
		Date  string `json:"date" binding:"required"`
		Shift int    `json:"shift" binding:"required"`
		Num   int    `json:"num" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Dữ liệu không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Validate date format (YYYY-MM-DD)
	_, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Định dạng ngày không hợp lệ, yêu cầu định dạng YYYY-MM-DD",
		})
		return
	}

	// Validate shift
	if req.Shift < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Ca làm việc không hợp lệ",
		})
		return
	}

	// Validate number of employees
	if req.Num < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Số lượng nhân viên phải lớn hơn 0",
		})
		return
	}

	// Check if a record already exists
	var limitEmployee models.LimitEmployee
	result := db.Where("date = ? AND shift = ?", req.Date, req.Shift).First(&limitEmployee)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create new record
			newLimit := models.LimitEmployee{
				Date:  req.Date,
				Shift: req.Shift,
				Num:   req.Num,
			}
			if err := db.Create(&newLimit).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Không thể tạo giới hạn nhân viên",
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": fmt.Sprintf("Đã giới hạn %d nhân viên cho ca %d vào ngày %s", req.Num, req.Shift, req.Date),
				"limit": gin.H{
					"id":    newLimit.ID,
					"date":  newLimit.Date,
					"shift": newLimit.Shift,
					"num":   newLimit.Num,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Lỗi khi kiểm tra giới hạn nhân viên",
			"error":   result.Error.Error(),
		})
		return
	}

	// Update existing record
	limitEmployee.Num = req.Num
	if err := db.Save(&limitEmployee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể cập nhật giới hạn nhân viên",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Đã cập nhật giới hạn %d nhân viên cho ca %d vào ngày %s", req.Num, req.Shift, req.Date),
		"limit": gin.H{
			"id":    limitEmployee.ID,
			"date":  limitEmployee.Date,
			"shift": limitEmployee.Shift,
			"num":   limitEmployee.Num,
		},
	})
}

// DeleteShiftLimit removes a shift limit
func (*Admin) DeleteShiftLimit(c *gin.Context) {
	db := global.Mdb

	// Get ID from URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID không hợp lệ",
		})
		return
	}

	// Find the limit
	var limitEmployee models.LimitEmployee
	if err := db.First(&limitEmployee, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Không tìm thấy giới hạn",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Lỗi khi truy xuất giới hạn",
				"error":   err.Error(),
			})
		}
		return
	}

	// Delete the limit
	if err := db.Delete(&limitEmployee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể xóa giới hạn",
			"error":   err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Đã xóa giới hạn cho ca %d vào ngày %s, sẽ sử dụng giới hạn mặc định", limitEmployee.Shift, limitEmployee.Date),
	})
}

// GetShiftLimits gets all shift limits
func (*Admin) GetShiftLimits(c *gin.Context) {
	db := global.Mdb

	// Query all shift limits
	var limits []models.LimitEmployee
	if err := db.Find(&limits).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể lấy danh sách giới hạn ca làm việc",
			"error":   err.Error(),
		})
		return
	}

	// Format response
	var formattedLimits []gin.H
	for _, limit := range limits {
		formattedLimits = append(formattedLimits, gin.H{
			"id":    limit.ID,
			"date":  limit.Date,
			"shift": limit.Shift,
			"num":   limit.Num,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Lấy danh sách giới hạn ca làm việc thành công",
		"limits":  formattedLimits,
		"default": DefaultShiftLimit,
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

func (*Admin) RejectTakeleave(c *gin.Context) {
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

	result := db.Model(&models.TakeLeave{}).Where("id = ?", body.ID).Update("is_agree", "NO")

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update take leave request",
		})
		return
	}

	// Phản hồi thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Take leave request rejected successfully",
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
	var body struct {
		IsActive string `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
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

	var newActiveStatus string
	var message string

	// Kiểm tra xem có phải yêu cầu toggle không
	if body.IsActive == "toggle" {
		// Truy vấn trạng thái hiện tại của nhân viên
		var employee models.Employee
		if err := db.First(&employee, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Không tìm thấy nhân viên",
			})
			return
		}

		// Chuyển đổi trạng thái
		if employee.IsActive == "YES" {
			newActiveStatus = "NO"
			message = "Đã vô hiệu hóa tài khoản nhân viên"
		} else {
			newActiveStatus = "YES"
			message = "Đã kích hoạt tài khoản nhân viên"
		}
	} else {
		// Sử dụng giá trị được gửi đến trực tiếp
		newActiveStatus = body.IsActive
		if newActiveStatus == "YES" {
			message = "Đã kích hoạt tài khoản nhân viên"
		} else {
			message = "Đã vô hiệu hóa tài khoản nhân viên"
		}
	}

	// Cập nhật trạng thái
	result := db.Model(&models.Employee{}).Where("id = ?", id).Update("is_active", newActiveStatus)

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
		"success": true,
		"message": message,
	})
}

// UpdateEmployee updates an employee's information
func (*Admin) UpdateEmployee(c *gin.Context) {
	// Lấy db
	db := global.Mdb

	// Lấy ID từ URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Định dạng ID không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Tìm nhân viên hiện tại
	var existingEmployee models.Employee
	if err := db.First(&existingEmployee, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Không tìm thấy nhân viên",
			"error":   err.Error(),
		})
		return
	}

	// Móc dữ liệu từ request body
	var updateData struct {
		FullName       string `json:"FullName"`
		Gender         string `json:"Gender"`
		Phone          string `json:"Phone"`
		Address        string `json:"Address"`
		DateOfBirth    string `json:"DateOfBirth"`
		HireDate       string `json:"HireDate"`
		Email          string `json:"Email"`
		Image          string `json:"Image"`
		CardNumber     string `json:"CardNumber"`
		Bank           string `json:"Bank"`
		Role           int    `json:"Role"`
		SalaryPartTime int    `json:"SalaryPartTime"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Dữ liệu không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// In ra dữ liệu nhận được để debug
	fmt.Printf("Received update data: %+v\n", updateData)

	// Cập nhật thông tin nhân viên
	updates := map[string]interface{}{
		"full_name":        updateData.FullName,
		"gender":           updateData.Gender,
		"phone":            updateData.Phone,
		"address":          updateData.Address,
		"date_of_birth":    updateData.DateOfBirth,
		"hire_date":        updateData.HireDate,
		"email":            updateData.Email,
		"image":            updateData.Image,
		"card_number":      updateData.CardNumber,
		"bank":             updateData.Bank,
		"role":             updateData.Role,
		"salary_part_time": updateData.SalaryPartTime,
	}

	// Chỉ cập nhật các trường không rỗng
	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if strValue, ok := value.(string); ok && strValue != "" {
			filteredUpdates[key] = value
		} else if intValue, ok := value.(int); ok && intValue != 0 {
			filteredUpdates[key] = value
		}
	}

	// Nếu không có gì để cập nhật, trả về thông báo
	if len(filteredUpdates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Không có thông tin nào được cập nhật",
		})
		return
	}

	result := db.Model(&models.Employee{}).Where("id = ?", id).Updates(filteredUpdates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Không thể cập nhật thông tin nhân viên",
			"error":   result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Không có thông tin nào được cập nhật",
		})
		return
	}

	// Trả về thông tin nhân viên đã cập nhật
	var updatedEmployee models.Employee
	db.First(&updatedEmployee, id)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Cập nhật thông tin nhân viên thành công",
		"employee": updatedEmployee,
	})
}

// DeleteSalary deletes a salary type by ID
func (*Admin) DeleteSalary(c *gin.Context) {
	// Lấy DB
	db := global.Mdb

	// Lấy ID từ URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Định dạng ID không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Kiểm tra xem có nhân viên nào đang sử dụng mức lương này không
	var count int64
	if err := db.Model(&models.Employee{}).Where("salary_part_time = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Lỗi khi kiểm tra dữ liệu nhân viên",
			"error":   err.Error(),
		})
		return
	}

	// Nếu có nhân viên đang sử dụng, trả về lỗi
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không thể xóa mức lương này vì có nhân viên đang sử dụng",
			"count":   count,
		})
		return
	}

	// Tìm kiếm mức lương để đảm bảo nó tồn tại
	var salary models.SalaryPartTime
	if err := db.First(&salary, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Không tìm thấy mức lương",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Lỗi khi tìm kiếm mức lương",
				"error":   err.Error(),
			})
		}
		return
	}

	// Thực hiện xóa mức lương
	result := db.Delete(&models.SalaryPartTime{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể xóa mức lương",
			"error":   result.Error.Error(),
		})
		return
	}

	// Trả về kết quả thành công
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Xóa mức lương thành công",
	})
}

func (*Admin) GetBonuses(c *gin.Context) {
	// Lấy db
	db := global.Mdb

	// Lấy tham số phân trang
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	// Chuyển đổi thành số nguyên
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	// Lấy employee_id từ query nếu có
	employeeID := c.Query("employee_id")

	// Tạo query
	query := db.Model(&models.Bonus{})

	// Nếu có employee_id, thêm điều kiện lọc
	if employeeID != "" {
		empID, err := strconv.Atoi(employeeID)
		if err == nil {
			query = query.Where("employee_id = ?", empID)
		}
	}

	// Lấy tổng số bản ghi
	var total int64
	query.Count(&total)

	// Lấy danh sách bonuses
	var bonuses []models.Bonus
	result := query.Limit(limitInt).Offset(offsetInt).Order("date DESC, time DESC").Find(&bonuses)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch bonuses",
		})
		return
	}

	// Chuẩn bị response với thông tin phân trang
	c.JSON(http.StatusOK, gin.H{
		"bonuses": bonuses,
		"pagination": gin.H{
			"total":  total,
			"limit":  limitInt,
			"offset": offsetInt,
		},
	})
}

// UpdateBonus updates a bonus by ID
func (*Admin) UpdateBonus(c *gin.Context) {
	// Lấy DB
	db := global.Mdb

	// Lấy ID từ URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Định dạng ID không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Tìm kiếm bonus để đảm bảo nó tồn tại
	var bonus models.Bonus
	if err := db.First(&bonus, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Không tìm thấy khoản thưởng",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Lỗi khi tìm kiếm khoản thưởng",
				"error":   err.Error(),
			})
		}
		return
	}

	// Kiểm tra xem bonus đã được thanh toán chưa
	if bonus.IsPayment == "OK" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không thể cập nhật khoản thưởng đã thanh toán",
		})
		return
	}

	// Móc dữ liệu từ request body
	var updateData struct {
		Date        string  `json:"Date"`
		Time        string  `json:"Time"`
		Description string  `json:"Description"`
		Money       float64 `json:"Money"`
		IsPayment   string  `json:"IsPayment"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Dữ liệu không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Cập nhật các trường có giá trị
	updates := map[string]interface{}{}

	if updateData.Date != "" {
		updates["date"] = updateData.Date
	}
	if updateData.Time != "" {
		updates["time"] = updateData.Time
	}
	if updateData.Description != "" {
		updates["description"] = updateData.Description
	}
	if updateData.Money > 0 {
		updates["money"] = updateData.Money
	}
	if updateData.IsPayment != "" {
		updates["is_payment"] = updateData.IsPayment
	}

	// Nếu không có gì để cập nhật, trả về thông báo
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không có thông tin nào được cập nhật",
		})
		return
	}

	// Thực hiện cập nhật
	result := db.Model(&models.Bonus{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể cập nhật khoản thưởng",
			"error":   result.Error.Error(),
		})
		return
	}

	// Trả về kết quả thành công
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cập nhật khoản thưởng thành công",
	})
}

// DeleteBonus deletes a bonus by ID
func (*Admin) DeleteBonus(c *gin.Context) {
	// Lấy DB
	db := global.Mdb

	// Lấy ID từ URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Định dạng ID không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Tìm kiếm bonus để đảm bảo nó tồn tại
	var bonus models.Bonus
	if err := db.First(&bonus, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Không tìm thấy khoản thưởng",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Lỗi khi tìm kiếm khoản thưởng",
				"error":   err.Error(),
			})
		}
		return
	}

	// Kiểm tra xem bonus đã được thanh toán chưa
	if bonus.IsPayment == "OK" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không thể xóa khoản thưởng đã thanh toán",
		})
		return
	}

	// Kiểm tra xem có payment_infor nào liên kết đến bonus này không
	var count int64
	if err := db.Model(&models.Payment_Infor{}).Where("bonus = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Lỗi khi kiểm tra dữ liệu thanh toán",
			"error":   err.Error(),
		})
		return
	}

	// Nếu có payment_infor liên kết, không cho phép xóa
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không thể xóa khoản thưởng đã liên kết với giao dịch thanh toán",
		})
		return
	}

	// Thực hiện xóa bonus
	result := db.Delete(&models.Bonus{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể xóa khoản thưởng",
			"error":   result.Error.Error(),
		})
		return
	}

	// Trả về kết quả thành công
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Xóa khoản thưởng thành công",
	})
}

// GetErrorTypes lấy danh sách các loại lỗi
func (*Admin) GetErrorTypes(c *gin.Context) {
	// Set cơ sở dữ liệu
	db := global.Mdb

	// Lấy danh sách loại lỗi
	var errorTypes []models.ErrorName
	if err := db.Find(&errorTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Không thể lấy danh sách loại lỗi",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"errorTypes": errorTypes,
	})
}

// UpdateErrorType cập nhật thông tin của một loại lỗi
func (*Admin) UpdateErrorType(c *gin.Context) {
	// Lấy DB
	db := global.Mdb

	// Lấy ID từ URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Định dạng ID không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Tìm kiếm error type để đảm bảo nó tồn tại
	var errorType models.ErrorName
	if err := db.First(&errorType, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Không tìm thấy loại lỗi",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Lỗi khi tìm kiếm loại lỗi",
				"error":   err.Error(),
			})
		}
		return
	}

	// Móc dữ liệu từ request body
	var updateData struct {
		NameError string  `json:"NameError"`
		Fines     float64 `json:"Fines"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Dữ liệu không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Cập nhật các trường có giá trị
	updates := map[string]interface{}{}

	if updateData.NameError != "" {
		updates["name_error"] = updateData.NameError
	}
	if updateData.Fines > 0 {
		updates["fines"] = updateData.Fines
	}

	// Nếu không có gì để cập nhật, trả về thông báo
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không có thông tin nào được cập nhật",
		})
		return
	}

	// Thực hiện cập nhật
	result := db.Model(&models.ErrorName{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể cập nhật loại lỗi",
			"error":   result.Error.Error(),
		})
		return
	}

	// Trả về kết quả thành công
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cập nhật loại lỗi thành công",
	})
}

// DeleteErrorType xóa một loại lỗi
func (*Admin) DeleteErrorType(c *gin.Context) {
	// Lấy DB
	db := global.Mdb

	// Lấy ID từ URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Định dạng ID không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Tìm kiếm error type để đảm bảo nó tồn tại
	var errorType models.ErrorName
	if err := db.First(&errorType, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Không tìm thấy loại lỗi",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Lỗi khi tìm kiếm loại lỗi",
				"error":   err.Error(),
			})
		}
		return
	}

	// Kiểm tra xem có Error (lỗi của nhân viên) nào đang sử dụng loại lỗi này không
	var count int64
	if err := db.Model(&models.Error{}).Where("type_error = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Lỗi khi kiểm tra dữ liệu lỗi nhân viên",
			"error":   err.Error(),
		})
		return
	}

	// Nếu có Error đang sử dụng, không cho phép xóa
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không thể xóa loại lỗi này vì đã có nhân viên bị ghi nhận lỗi này",
		})
		return
	}

	// Thực hiện xóa ErrorName
	result := db.Delete(&models.ErrorName{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể xóa loại lỗi",
			"error":   result.Error.Error(),
		})
		return
	}

	// Trả về kết quả thành công
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Xóa loại lỗi thành công",
	})
}

// GetEmployeeErrors lấy danh sách lỗi đã ghi nhận cho tất cả nhân viên hoặc một nhân viên cụ thể
func (*Admin) GetEmployeeErrors(c *gin.Context) {
	// Lấy DB
	db := global.Mdb

	// Lấy tham số employeeID nếu có
	employeeIDStr := c.Query("employee_id")
	var employeeID int
	var err error
	if employeeIDStr != "" {
		employeeID, err = strconv.Atoi(employeeIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "employee_id không hợp lệ",
				"error":   err.Error(),
			})
			return
		}
	}

	// Lấy tham số phân trang
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Tạo query cơ bản
	query := db.Model(&models.Error{}).
		Select("errors.*, error_names.name_error, error_names.fines").
		Joins("LEFT JOIN error_names ON errors.type_error = error_names.id").
		Order("errors.date DESC, errors.time DESC")

	// Thêm điều kiện lọc theo employeeID nếu có
	if employeeIDStr != "" {
		query = query.Where("errors.employee_id = ?", employeeID)
	}

	// Đếm tổng số bản ghi
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Không thể đếm số lỗi",
			"error":   err.Error(),
		})
		return
	}

	// Thực hiện truy vấn với phân trang
	var errors []struct {
		models.Error
		NameError string  `json:"nameError"`
		Fines     float64 `json:"fines"`
	}

	if err := query.Limit(limit).Offset(offset).Find(&errors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Không thể lấy danh sách lỗi",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   errors,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (*Admin) GetUnpaidAttendance(c *gin.Context) {
	// Get the database connection
	db := global.Mdb

	// Get the employee ID from the URL parameter
	employeeIDStr := c.Param("id")
	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	// Query all attendance records for this employee with both check-in and check-out times
	var attendances []models.Attendance
	if err := db.Where("employee_id = ? AND check_in != '' AND check_out != ''", employeeID).Find(&attendances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying attendance records"})
		return
	}

	// If no attendance records found, return an empty array
	if len(attendances) == 0 {
		c.JSON(http.StatusOK, gin.H{"attendances": []models.Attendance{}})
		return
	}

	// Extract the IDs of all attendance records
	var attendanceIDs []uint
	for _, attendance := range attendances {
		attendanceIDs = append(attendanceIDs, attendance.ID)
	}

	// Query payment_infor table to find which attendance IDs are already paid
	var paymentInfos []models.Payment_Infor
	if err := db.Where("attendance_id IN ?", attendanceIDs).Find(&paymentInfos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying payment info records"})
		return
	}

	// Create a map of paid attendance IDs for quick lookup
	paidAttendanceIDs := make(map[uint]bool)
	for _, paymentInfo := range paymentInfos {
		if paymentInfo.AttendanceID > 0 {
			paidAttendanceIDs[uint(paymentInfo.AttendanceID)] = true
		}
	}

	// Filter out the attendance records that have already been paid
	var unpaidAttendances []models.Attendance
	for _, attendance := range attendances {
		if !paidAttendanceIDs[attendance.ID] {
			unpaidAttendances = append(unpaidAttendances, attendance)
		}
	}

	// Return the unpaid attendance records
	c.JSON(http.StatusOK, gin.H{"attendances": unpaidAttendances})
}

func (*Admin) GetPaymentInfo(c *gin.Context) {
	// Get the database connection
	db := global.Mdb

	// Get the employee ID from the URL parameter
	employeeIDStr := c.Param("id")
	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	// First, get all payments for this employee
	var payments []models.Payment
	if err := db.Where("employee_id = ?", employeeID).Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying payment records"})
		return
	}

	// Extract payment IDs
	var paymentIDs []uint
	for _, payment := range payments {
		paymentIDs = append(paymentIDs, payment.ID)
	}

	// If no payments found, return an empty array
	if len(paymentIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// Get all payment info records related to these payments
	var paymentInfos []models.Payment_Infor
	if err := db.Where("id_payment IN ?", paymentIDs).Find(&paymentInfos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying payment info records"})
		return
	}

	// Return the payment info records
	c.JSON(http.StatusOK, paymentInfos)
}

func (*Admin) GetUnpaidBonuses(c *gin.Context) {
	// Get the database connection
	db := global.Mdb

	// Get the employee ID from the URL parameter
	employeeIDStr := c.Param("id")
	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	// Query all bonus records for this employee that are not marked as paid
	var bonuses []models.Bonus
	if err := db.Where("employee_id = ? AND is_payment = 'NO'", employeeID).Find(&bonuses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying bonus records"})
		return
	}

	// If no bonus records found, return an empty array
	if len(bonuses) == 0 {
		c.JSON(http.StatusOK, gin.H{"bonuses": []models.Bonus{}})
		return
	}

	// Extract the IDs of all bonus records
	var bonusIDs []uint
	for _, bonus := range bonuses {
		bonusIDs = append(bonusIDs, bonus.ID)
	}

	// Query payment_infor table to find which bonus IDs are already paid
	var paymentInfos []models.Payment_Infor
	if err := db.Where("bonus IN ?", bonusIDs).Find(&paymentInfos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying payment info records"})
		return
	}

	// Create a map of paid bonus IDs for quick lookup
	paidBonusIDs := make(map[uint]bool)
	for _, paymentInfo := range paymentInfos {
		if paymentInfo.Bonus > 0 {
			paidBonusIDs[uint(paymentInfo.Bonus)] = true
		}
	}

	// Filter out the bonus records that have already been paid
	var unpaidBonuses []models.Bonus
	for _, bonus := range bonuses {
		if !paidBonusIDs[bonus.ID] {
			unpaidBonuses = append(unpaidBonuses, bonus)
		}
	}

	// Return the unpaid bonus records
	c.JSON(http.StatusOK, gin.H{"bonuses": unpaidBonuses})
}

func (*Admin) GetUnpaidErrors(c *gin.Context) {
	// Get the database connection
	db := global.Mdb

	// Get the employee ID from the URL parameter
	employeeIDStr := c.Param("id")
	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	// Query all error records for this employee that are not marked as paid
	var errors []models.Error
	if err := db.Where("employee_id = ? AND is_payment = 'NO'", employeeID).Find(&errors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying error records"})
		return
	}

	// If no error records found, return an empty array
	if len(errors) == 0 {
		c.JSON(http.StatusOK, gin.H{"errors": []models.Error{}})
		return
	}

	// Extract the IDs of all error records
	var errorIDs []uint
	for _, err := range errors {
		errorIDs = append(errorIDs, err.ID)
	}

	// Query payment_infor table to find which error IDs are already paid
	var paymentInfos []models.Payment_Infor
	if err := db.Where("error IN ?", errorIDs).Find(&paymentInfos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying payment info records"})
		return
	}

	// Create a map of paid error IDs for quick lookup
	paidErrorIDs := make(map[uint]bool)
	for _, paymentInfo := range paymentInfos {
		if paymentInfo.Error > 0 {
			paidErrorIDs[uint(paymentInfo.Error)] = true
		}
	}

	// Filter out the error records that have already been paid
	var unpaidErrors []models.Error
	for _, err := range errors {
		if !paidErrorIDs[err.ID] {
			unpaidErrors = append(unpaidErrors, err)
		}
	}

	// Return the unpaid error records
	c.JSON(http.StatusOK, gin.H{"errors": unpaidErrors})
}

// GetPayments retrieves the payment history with optional filtering
func (*Admin) GetPayments(c *gin.Context) {
	// Get the database connection
	db := global.Mdb

	// Parse query parameters
	employeeIDStr := c.Query("employee_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	// Convert to appropriate types
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Build the base query
	query := db.Model(&models.Payment{}).
		Select("payments.*, employees.full_name as employee_name").
		Joins("LEFT JOIN employees ON payments.employee_id = employees.id")

	// Apply filters if provided
	if employeeIDStr != "" {
		employeeID, err := strconv.Atoi(employeeIDStr)
		if err == nil {
			query = query.Where("payments.employee_id = ?", employeeID)
		}
	}

	if startDate != "" {
		query = query.Where("payments.date >= ?", startDate)
	}

	if endDate != "" {
		query = query.Where("payments.date <= ?", endDate)
	}

	// Count total records for pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error counting payment records",
		})
		return
	}

	// Execute the query with pagination
	type PaymentWithName struct {
		models.Payment
		EmployeeName    string  `json:"employeeName"`
		TotalIncome     float64 `json:"totalIncome" gorm:"-"`
		TotalDeductions float64 `json:"totalDeductions" gorm:"-"`
		TotalAmount     float64 `json:"totalAmount" gorm:"-"`
	}

	var payments []PaymentWithName
	if err := query.Order("payments.date DESC, payments.time DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error retrieving payment records",
		})
		return
	}

	// Calculate totals for each payment
	for i := range payments {
		// Get employee salary info
		var employee models.Employee
		if err := db.First(&employee, payments[i].EmployeeID).Error; err != nil {
			continue
		}

		// Get salary details
		var salary models.SalaryPartTime
		if err := db.First(&salary, employee.SalaryPartTime).Error; err != nil {
			continue
		}

		// Get payment details: attendance, bonuses, errors
		var paymentInfos []models.Payment_Infor
		if err := db.Where("id_payment = ?", payments[i].ID).Find(&paymentInfos).Error; err != nil {
			continue
		}

		// Extract IDs for each type
		var attendanceIDs, bonusIDs, errorIDs []int
		for _, info := range paymentInfos {
			if info.AttendanceID > 0 {
				attendanceIDs = append(attendanceIDs, info.AttendanceID)
			}
			if info.Bonus > 0 {
				bonusIDs = append(bonusIDs, info.Bonus)
			}
			if info.Error > 0 {
				errorIDs = append(errorIDs, info.Error)
			}
		}

		// Calculate income from attendance
		totalIncome := 0.0
		if len(attendanceIDs) > 0 {
			var attendances []models.Attendance
			if err := db.Where("id IN ?", attendanceIDs).Find(&attendances).Error; err == nil {
				for _, attendance := range attendances {
					// Calculate work hours
					checkIn, _ := time.Parse("15:04:05", attendance.CheckIn)
					checkOut, _ := time.Parse("15:04:05", attendance.CheckOut)
					hours := checkOut.Sub(checkIn).Hours()

					// Add to total income
					totalIncome += hours * salary.Salary
				}
			}
		}

		// Add bonuses to income
		if len(bonusIDs) > 0 {
			var bonuses []models.Bonus
			if err := db.Where("id IN ?", bonusIDs).Find(&bonuses).Error; err == nil {
				for _, bonus := range bonuses {
					totalIncome += bonus.Money
				}
			}
		}

		// Calculate deductions from errors
		totalDeductions := 0.0
		if len(errorIDs) > 0 {
			var errors []struct {
				ID        uint
				TypeError int
			}
			if err := db.Model(&models.Error{}).Where("id IN ?", errorIDs).Find(&errors).Error; err == nil {
				for _, err := range errors {
					// Get the fine amount from error type
					var errorType models.ErrorName
					if dbErr := db.First(&errorType, err.TypeError).Error; dbErr == nil {
						totalDeductions += errorType.Fines
					}
				}
			}
		}

		// Set the calculated values
		payments[i].TotalIncome = totalIncome
		payments[i].TotalDeductions = totalDeductions
		payments[i].TotalAmount = totalIncome - totalDeductions
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetRegistrations returns all shift registrations
func (*Admin) GetRegistrations(c *gin.Context) {
	db := global.Mdb

	// Get query parameters for filtering
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")

	// Parse limit and offset
	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	// Get registrations with pagination
	var registrations []models.Registration
	var total int64

	// Get total count
	db.Model(&models.Registration{}).Count(&total)

	// Get data with pagination
	query := db.Model(&models.Registration{}).Offset(offsetInt).Limit(limitInt).Order("date DESC, shift ASC")
	if err := query.Find(&registrations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể truy xuất danh sách đăng ký",
			"error":   err.Error(),
		})
		return
	}

	// Process registrations to include employee names if needed
	var processedRegistrations []map[string]interface{}
	for _, reg := range registrations {
		// Get employee information
		var employee models.Employee
		if err := db.First(&employee, reg.EmployeeID).Error; err == nil {
			// Employee found, add their name
			processedRegistrations = append(processedRegistrations, map[string]interface{}{
				"id":           reg.ID,
				"employeeId":   reg.EmployeeID,
				"employeeName": employee.FullName,
				"date":         reg.Date,
				"shift":        reg.Shift,
				"createdAt":    reg.CreatedAt,
				"updatedAt":    reg.UpdatedAt,
			})
		} else {
			// Employee not found, still add the registration
			processedRegistrations = append(processedRegistrations, map[string]interface{}{
				"id":           reg.ID,
				"employeeId":   reg.EmployeeID,
				"employeeName": "Không xác định",
				"date":         reg.Date,
				"shift":        reg.Shift,
				"createdAt":    reg.CreatedAt,
				"updatedAt":    reg.UpdatedAt,
			})
		}
	}

	// Return data with pagination info
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"registrations": processedRegistrations,
		"pagination": gin.H{
			"total":  total,
			"limit":  limitInt,
			"offset": offsetInt,
		},
	})
}

// GetRegistrationByID returns a specific registration by ID
func (*Admin) GetRegistrationByID(c *gin.Context) {
	db := global.Mdb

	// Get ID from URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID không hợp lệ",
		})
		return
	}

	// Find registration
	var registration models.Registration
	if err := db.First(&registration, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Không tìm thấy đăng ký",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Lỗi khi truy xuất đăng ký",
				"error":   err.Error(),
			})
		}
		return
	}

	// Get employee information
	var employee models.Employee
	employeeName := "Không xác định"
	if err := db.First(&employee, registration.EmployeeID).Error; err == nil {
		employeeName = employee.FullName
	}

	// Return registration with employee name
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"registration": gin.H{
			"id":           registration.ID,
			"employeeId":   registration.EmployeeID,
			"employeeName": employeeName,
			"date":         registration.Date,
			"shift":        registration.Shift,
			"createdAt":    registration.CreatedAt,
			"updatedAt":    registration.UpdatedAt,
		},
	})
}

// GetRegistrationsByEmployee returns all registrations for a specific employee
func (*Admin) GetRegistrationsByEmployee(c *gin.Context) {
	db := global.Mdb

	// Get employee ID from URL
	employeeIDStr := c.Param("id")
	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID nhân viên không hợp lệ",
		})
		return
	}

	// Get query parameters for pagination
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")

	// Parse limit and offset
	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	// Get employee information
	var employee models.Employee
	employeeName := "Không xác định"
	if err := db.First(&employee, employeeID).Error; err == nil {
		employeeName = employee.FullName
	} else {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Không tìm thấy nhân viên",
			})
			return
		}
	}

	// Get registrations with pagination
	var registrations []models.Registration
	var total int64

	// Get total count
	db.Model(&models.Registration{}).Where("employee_id = ?", employeeID).Count(&total)

	// Get data with pagination
	query := db.Model(&models.Registration{}).Where("employee_id = ?", employeeID).Offset(offsetInt).Limit(limitInt).Order("date DESC, shift ASC")
	if err := query.Find(&registrations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể truy xuất danh sách đăng ký",
			"error":   err.Error(),
		})
		return
	}

	// Process registrations
	var processedRegistrations []map[string]interface{}
	for _, reg := range registrations {
		processedRegistrations = append(processedRegistrations, map[string]interface{}{
			"id":           reg.ID,
			"employeeId":   reg.EmployeeID,
			"employeeName": employeeName,
			"date":         reg.Date,
			"shift":        reg.Shift,
			"createdAt":    reg.CreatedAt,
			"updatedAt":    reg.UpdatedAt,
		})
	}

	// Return data with pagination info
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"employee": gin.H{
			"id":   employeeID,
			"name": employeeName,
		},
		"registrations": processedRegistrations,
		"pagination": gin.H{
			"total":  total,
			"limit":  limitInt,
			"offset": offsetInt,
		},
	})
}

// GetRegistrationsByDate returns all registrations for a specific date
func (*Admin) GetRegistrationsByDate(c *gin.Context) {
	db := global.Mdb

	// Get date from URL
	date := c.Param("date")

	// Validate date format (should be YYYY-MM-DD)
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Ngày không hợp lệ, sử dụng định dạng YYYY-MM-DD",
		})
		return
	}

	// Get registrations for the specified date
	var registrations []models.Registration
	if err := db.Where("date = ?", date).Order("shift ASC").Find(&registrations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể truy xuất danh sách đăng ký",
			"error":   err.Error(),
		})
		return
	}

	// Process registrations to include employee names
	var processedRegistrations []map[string]interface{}
	for _, reg := range registrations {
		// Get employee information
		var employee models.Employee
		employeeName := "Không xác định"
		if err := db.First(&employee, reg.EmployeeID).Error; err == nil {
			employeeName = employee.FullName
		}

		processedRegistrations = append(processedRegistrations, map[string]interface{}{
			"id":           reg.ID,
			"employeeId":   reg.EmployeeID,
			"employeeName": employeeName,
			"date":         reg.Date,
			"shift":        reg.Shift,
			"createdAt":    reg.CreatedAt,
			"updatedAt":    reg.UpdatedAt,
		})
	}

	// Get shift counts by shift number
	shiftCounts := make(map[int]int)
	for _, reg := range registrations {
		shiftCounts[reg.Shift]++
	}

	// Return data
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"date":          date,
		"registrations": processedRegistrations,
		"shiftCounts":   shiftCounts,
		"total":         len(processedRegistrations),
	})
}

// GetRegistrationsByShift returns all registrations for a specific shift
func (*Admin) GetRegistrationsByShift(c *gin.Context) {
	db := global.Mdb

	// Get shift from URL
	shiftStr := c.Param("shift")
	shift, err := strconv.Atoi(shiftStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Mã ca làm việc không hợp lệ",
		})
		return
	}

	// Get query parameters for filtering
	date := c.DefaultQuery("date", "")

	// Build query
	query := db.Model(&models.Registration{}).Where("shift = ?", shift)

	// Add date filter if provided
	if date != "" {
		// Validate date format
		_, err := time.Parse("2006-01-02", date)
		if err == nil {
			query = query.Where("date = ?", date)
		}
	}

	// Get registrations
	var registrations []models.Registration
	if err := query.Order("date DESC").Find(&registrations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể truy xuất danh sách đăng ký",
			"error":   err.Error(),
		})
		return
	}

	// Process registrations to include employee names
	var processedRegistrations []map[string]interface{}
	for _, reg := range registrations {
		// Get employee information
		var employee models.Employee
		employeeName := "Không xác định"
		if err := db.First(&employee, reg.EmployeeID).Error; err == nil {
			employeeName = employee.FullName
		}

		processedRegistrations = append(processedRegistrations, map[string]interface{}{
			"id":           reg.ID,
			"employeeId":   reg.EmployeeID,
			"employeeName": employeeName,
			"date":         reg.Date,
			"shift":        reg.Shift,
			"createdAt":    reg.CreatedAt,
			"updatedAt":    reg.UpdatedAt,
		})
	}

	// Return data
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"shift":         shift,
		"date":          date, // Will be empty string if not provided
		"registrations": processedRegistrations,
		"total":         len(processedRegistrations),
	})
}

// CreateRegistration creates a new registration
func (*Admin) CreateRegistration(c *gin.Context) {
	db := global.Mdb

	// Parse request body
	var requestBody struct {
		EmployeeID int64  `json:"employeeId"`
		Date       string `json:"date"`
		Shift      int    `json:"shift"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Dữ liệu không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Validate employee ID
	var employee models.Employee
	if err := db.First(&employee, requestBody.EmployeeID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Nhân viên không tồn tại",
		})
		return
	}

	// Validate date format
	_, err := time.Parse("2006-01-02", requestBody.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Ngày không hợp lệ, sử dụng định dạng YYYY-MM-DD",
		})
		return
	}

	// Validate shift
	if requestBody.Shift < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Mã ca làm việc không hợp lệ",
		})
		return
	}

	// Check if registration already exists
	var existingCount int64
	db.Model(&models.Registration{}).
		Where("employee_id = ? AND date = ? AND shift = ?", requestBody.EmployeeID, requestBody.Date, requestBody.Shift).
		Count(&existingCount)

	if existingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Nhân viên đã đăng ký ca làm việc này",
		})
		return
	}

	// Kiểm tra giới hạn nhân viên trong ca làm việc
	// 1. Kiểm tra xem có giới hạn cụ thể cho ngày và ca này không
	var limitEmployee models.LimitEmployee
	var shiftLimit int = DefaultShiftLimit // Mặc định là 6

	err = db.Where("date = ? AND shift = ?", requestBody.Date, requestBody.Shift).First(&limitEmployee).Error
	if err == nil {
		// Có giới hạn cụ thể trong cơ sở dữ liệu
		shiftLimit = limitEmployee.Num
	}

	// 2. Đếm số nhân viên đã đăng ký ca này
	var registeredCount int64
	db.Model(&models.Registration{}).
		Where("date = ? AND shift = ?", requestBody.Date, requestBody.Shift).
		Count(&registeredCount)

	// 3. So sánh với giới hạn
	if int(registeredCount) >= shiftLimit {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"message":        fmt.Sprintf("Ca làm việc này đã đạt giới hạn %d nhân viên", shiftLimit),
			"currentLimit":   shiftLimit,
			"currentCount":   registeredCount,
			"isDefaultLimit": err != nil, // true nếu đang sử dụng giới hạn mặc định
		})
		return
	}

	// Create the registration
	registration := models.Registration{
		EmployeeID: requestBody.EmployeeID,
		Date:       requestBody.Date,
		Shift:      requestBody.Shift,
	}

	if err := db.Create(&registration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể tạo đăng ký",
			"error":   err.Error(),
		})
		return
	}

	// Return success response with employee info
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Tạo đăng ký thành công",
		"registration": gin.H{
			"id":           registration.ID,
			"employeeId":   registration.EmployeeID,
			"employeeName": employee.FullName,
			"date":         registration.Date,
			"shift":        registration.Shift,
			"createdAt":    registration.CreatedAt,
			"updatedAt":    registration.UpdatedAt,
		},
		"shiftLimit":      shiftLimit,
		"registeredCount": registeredCount + 1,
		"isDefaultLimit":  err != nil, // true nếu đang sử dụng giới hạn mặc định
	})
}

// UpdateRegistration updates an existing registration
func (*Admin) UpdateRegistration(c *gin.Context) {
	db := global.Mdb

	// Get ID from URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID không hợp lệ",
		})
		return
	}

	// Parse request body
	var requestBody struct {
		EmployeeID int64  `json:"employeeId"`
		Date       string `json:"date"`
		Shift      int    `json:"shift"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Dữ liệu không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Find the registration
	var registration models.Registration
	if err := db.First(&registration, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Không tìm thấy đăng ký",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Lỗi khi truy xuất đăng ký",
				"error":   err.Error(),
			})
		}
		return
	}

	// Validate employee ID
	var employee models.Employee
	if err := db.First(&employee, requestBody.EmployeeID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Nhân viên không tồn tại",
		})
		return
	}

	// Validate date format
	_, err = time.Parse("2006-01-02", requestBody.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Ngày không hợp lệ, sử dụng định dạng YYYY-MM-DD",
		})
		return
	}

	// Validate shift
	if requestBody.Shift < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Mã ca làm việc không hợp lệ",
		})
		return
	}

	// Check if another registration with the same employee, date, and shift exists
	var existingCount int64
	db.Model(&models.Registration{}).
		Where("employee_id = ? AND date = ? AND shift = ? AND id != ?",
			requestBody.EmployeeID, requestBody.Date, requestBody.Shift, id).
		Count(&existingCount)

	if existingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Nhân viên đã đăng ký ca làm việc này",
		})
		return
	}

	// Update the registration
	updates := map[string]interface{}{
		"employee_id": requestBody.EmployeeID,
		"date":        requestBody.Date,
		"shift":       requestBody.Shift,
	}

	if err := db.Model(&registration).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể cập nhật đăng ký",
			"error":   err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cập nhật đăng ký thành công",
		"registration": gin.H{
			"id":           registration.ID,
			"employeeId":   registration.EmployeeID,
			"employeeName": employee.FullName,
			"date":         requestBody.Date,
			"shift":        requestBody.Shift,
			"updatedAt":    registration.UpdatedAt,
		},
	})
}

// DeleteRegistration deletes a registration
func (*Admin) DeleteRegistration(c *gin.Context) {
	db := global.Mdb

	// Get ID from URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID không hợp lệ",
		})
		return
	}

	// Find the registration
	var registration models.Registration
	if err := db.First(&registration, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Không tìm thấy đăng ký",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Lỗi khi truy xuất đăng ký",
				"error":   err.Error(),
			})
		}
		return
	}

	// Delete the registration
	if err := db.Delete(&registration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể xóa đăng ký",
			"error":   err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Xóa đăng ký thành công",
	})
}

// GetPaymentDetails retrieves detailed information about a specific payment
func (*Admin) GetPaymentDetails(c *gin.Context) {
	// Get the database connection
	db := global.Mdb

	// Get the payment ID from the URL parameter
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Payment ID is required",
		})
		return
	}

	// Get the payment record
	var payment models.Payment
	if err := db.First(&payment, paymentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Payment not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Error retrieving payment",
				"error":   err.Error(),
			})
		}
		return
	}

	// Get employee information
	var employee models.Employee
	if err := db.First(&employee, payment.EmployeeID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error retrieving employee information",
			"error":   err.Error(),
		})
		return
	}

	// Get salary information
	var salary models.SalaryPartTime
	salaryRate := 0.0
	if employee.SalaryPartTime > 0 {
		if err := db.First(&salary, employee.SalaryPartTime).Error; err == nil {
			salaryRate = salary.Salary
		}
	}

	// Get payment info records
	var paymentInfos []models.Payment_Infor
	if err := db.Where("id_payment = ?", payment.ID).Find(&paymentInfos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error retrieving payment information",
			"error":   err.Error(),
		})
		return
	}

	// Extract attendance, bonus, and error IDs
	var attendanceIDs, bonusIDs, errorIDs []uint
	for _, info := range paymentInfos {
		if info.AttendanceID > 0 {
			attendanceIDs = append(attendanceIDs, uint(info.AttendanceID))
		}
		if info.Bonus > 0 {
			bonusIDs = append(bonusIDs, uint(info.Bonus))
		}
		if info.Error > 0 {
			errorIDs = append(errorIDs, uint(info.Error))
		}
	}

	// Get attendance records with work hours
	var attendances []gin.H
	if len(attendanceIDs) > 0 {
		var attendanceModels []models.Attendance
		if err := db.Where("id IN ?", attendanceIDs).Find(&attendanceModels).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Error retrieving attendance records",
				"error":   err.Error(),
			})
			return
		}

		// Calculate work hours for each attendance and create proper response format
		for _, att := range attendanceModels {
			checkIn, _ := time.Parse("15:04:05", att.CheckIn)
			checkOut, _ := time.Parse("15:04:05", att.CheckOut)
			workHours := checkOut.Sub(checkIn).Hours()

			attendances = append(attendances, gin.H{
				"id":         att.ID,
				"employeeId": att.EmployeeID,
				"date":       att.Date,
				"checkIn":    att.CheckIn,
				"checkOut":   att.CheckOut,
				"shift":      att.Shift,
				"workHours":  workHours,
			})
		}
	}

	// Get bonus records
	var bonuses []gin.H
	if len(bonusIDs) > 0 {
		var bonusModels []models.Bonus
		if err := db.Where("id IN ?", bonusIDs).Find(&bonusModels).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Error retrieving bonus records",
				"error":   err.Error(),
			})
			return
		}

		for _, b := range bonusModels {
			bonuses = append(bonuses, gin.H{
				"id":          b.ID,
				"employeeId":  b.EmployeeID,
				"date":        b.Date,
				"time":        b.Time,
				"description": b.Description,
				"money":       b.Money,
				"isPayment":   b.IsPayment,
			})
		}
	}

	// Get error records with error names and fines
	var errors []gin.H
	if len(errorIDs) > 0 {
		type ErrorWithDetails struct {
			ID         uint
			EmployeeID int
			Date       string
			Time       string
			TypeError  int
			IsPayment  string
			Evidence   string
			NameError  string
			Fines      float64
		}

		var errorResults []ErrorWithDetails
		query := `
			SELECT e.id, e.employee_id, e.date, e.time, e.type_error, e.is_payment, e.evidence, 
			       en.name_error, en.fines 
			FROM errors e
			LEFT JOIN error_names en ON e.type_error = en.id
			WHERE e.id IN ?
		`

		if err := db.Raw(query, errorIDs).Scan(&errorResults).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Error retrieving error records",
				"error":   err.Error(),
			})
			return
		}

		for _, e := range errorResults {
			errors = append(errors, gin.H{
				"id":         e.ID,
				"employeeId": e.EmployeeID,
				"date":       e.Date,
				"time":       e.Time,
				"typeError":  e.TypeError,
				"isPayment":  e.IsPayment,
				"evidence":   e.Evidence,
				"nameError":  e.NameError,
				"fines":      e.Fines,
			})
		}
	}

	// Calculate payment totals
	var totalIncome, totalDeductions float64

	// Calculate income from attendance
	for _, att := range attendances {
		if workHours, ok := att["workHours"].(float64); ok {
			totalIncome += salaryRate * workHours
		}
	}

	// Add bonuses to income
	for _, bonus := range bonuses {
		if money, ok := bonus["money"].(float64); ok {
			totalIncome += money
		}
	}

	// Add deductions for errors
	for _, err := range errors {
		if fines, ok := err["fines"].(float64); ok {
			totalDeductions += fines
		}
	}

	// Format the response to match the frontend's expectations
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"payment": gin.H{
			"id":              payment.ID,
			"employeeId":      payment.EmployeeID,
			"employeeName":    employee.FullName,
			"date":            payment.Date,
			"time":            payment.Time,
			"evidence":        payment.Evidence,
			"totalIncome":     totalIncome,
			"totalDeductions": totalDeductions,
			"totalAmount":     totalIncome - totalDeductions,
		},
		"employee":        employee,
		"attendance":      attendances,
		"bonuses":         bonuses,
		"errors":          errors,
		"totalIncome":     totalIncome,
		"totalDeductions": totalDeductions,
		"totalAmount":     totalIncome - totalDeductions,
	})
}
