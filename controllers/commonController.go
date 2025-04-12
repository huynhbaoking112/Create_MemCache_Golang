package controllers

import (
	"fmt"
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

	// Send it back with user information
	c.JSON(http.StatusOK, gin.H{
		"message": "Login success",
		"user": gin.H{
			"id":             user.ID,
			"fullName":       user.FullName,
			"email":          user.Email,
			"phone":          user.Phone,
			"role":           user.Role,
			"address":        user.Address,
			"gender":         user.Gender,
			"dateOfBirth":    user.DateOfBirth,
			"hireDate":       user.HireDate,
			"image":          user.Image,
			"cardNumber":     user.CardNumber,
			"bank":           user.Bank,
			"newNOTI":        user.NewNOTI,
			"isActive":       user.IsActive,
			"salaryPartTime": user.SalaryPartTime,
			"createdAt":      user.CreatedAt,
			"token":          tokenString,
		},
	})
}

func (*Common) GetAttendance(c *gin.Context) {
	// lấy db:
	db := global.Mdb

	// Lấy thông tin user
	// user, exits := c.Get("user")

	// if !exits {
	// 	c.JSON(http.StatusForbidden, gin.H{
	// 		"message": "Bạn phải đăng nhập",
	// 	})
	// 	return
	// }

	// // Ép kiểu dữ liệu
	// userModel, ok := user.(models.Employee)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"message": "Failed to parse user data",
	// 	})
	// 	return
	// }

	// Lấy giá trị id
	idStr := c.Param("id")
	// Ép kiểu id
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Kiểm tra quyền
	// if userModel.Role != 2 && int(userModel.ID) != id {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền này"})
	// 	return
	// }

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

	// // Xử lý user
	// user, exists := c.Get("user")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Bạn phải đăng nhập"})
	// 	return
	// }
	// userModel, ok := user.(models.Employee)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi ép kiểu dữ liệu"})
	// 	return
	// }

	// Kiểm tra quyền
	// if userModel.ID != uint(id) && userModel.Role != 2 {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền này"})
	// 	return
	// }

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
	// user, exists := c.Get("user")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Bạn phải đăng nhập"})
	// 	return
	// }
	// userModel, ok := user.(models.Employee)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi ép kiểu dữ liệu"})
	// 	return
	// }

	// Kiểm tra quyền
	// if userModel.ID != uint(id) && userModel.Role != 2 {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền này"})
	// 	return
	// }

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

// GetSalaries returns all salary types
func (cmm *Common) GetSalaries(c *gin.Context) {
	db := global.Mdb

	// Get all salary types
	var salaries []models.SalaryPartTime
	result := db.Find(&salaries)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch salary types",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"salaries": salaries,
	})
}

// GetEmployee returns a specific employee with their salary info
func (cmm *Common) GetEmployee(c *gin.Context) {
	db := global.Mdb

	// Get the employee ID from the URL
	employeeID := c.Param("id")

	// Define response struct
	type EmployeeResponse struct {
		Employee     models.Employee       `json:"employee"`
		Salary       models.SalaryPartTime `json:"salary,omitempty"`
		SalaryAmount float64               `json:"salaryAmount"`
	}

	response := EmployeeResponse{
		SalaryAmount: 0, // Default to 0
	}

	// Query the employee
	if err := db.First(&response.Employee, employeeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Employee not found",
		})
		return
	}

	// Get the salary info if employee has a salary type
	if response.Employee.SalaryPartTime > 0 {
		if err := db.First(&response.Salary, response.Employee.SalaryPartTime).Error; err == nil {
			response.SalaryAmount = response.Salary.Salary
		}
	}

	// Return the response
	c.JSON(http.StatusOK, response)
}

// GetEmployees returns all employees with optional filtering
func (cmm *Common) GetEmployees(c *gin.Context) {
	db := global.Mdb

	// Get query parameters for filtering
	role := c.Query("role")
	isActive := c.Query("is_active")

	// Start building the query
	query := db

	// Apply filters if they exist
	if role != "" {
		query = query.Where("role = ?", role)
	}

	if isActive != "" {
		query = query.Where("is_active = ?", isActive)
	}

	// Get all employees
	var employees []models.Employee
	result := query.Find(&employees)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch employees",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"employees": employees,
	})
}

// GetEmployeesWithSalary returns all employees with their salary information
func (cmm *Common) GetEmployeesWithSalary(c *gin.Context) {
	db := global.Mdb

	// Get query parameters for filtering
	role := c.Query("role")
	isActive := c.Query("is_active")

	// Get employees
	var employees []models.Employee
	query := db

	// Apply filters if they exist
	if role != "" {
		query = query.Where("role = ?", role)
	}

	if isActive != "" {
		query = query.Where("is_active = ?", isActive)
	}

	if err := query.Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch employees",
		})
		return
	}

	// Create a slice to hold employee responses with salary data
	var employeesWithSalary []map[string]interface{}

	// For each employee, get their salary
	for _, emp := range employees {
		empData := map[string]interface{}{
			"ID":          emp.ID,
			"Email":       emp.Email,
			"FullName":    emp.FullName,
			"Gender":      emp.Gender,
			"Phone":       emp.Phone,
			"Address":     emp.Address,
			"DateOfBirth": emp.DateOfBirth,
			"HireDate":    emp.HireDate,
			"CardNumber":  emp.CardNumber,
			"Bank":        emp.Bank,
			"NewNOTI":     emp.NewNOTI,
			"IsActive":    emp.IsActive,
			"Role":        emp.Role,
			"Image":       emp.Image,
			"CreatedAt":   emp.CreatedAt,
			"UpdatedAt":   emp.UpdatedAt,
			// Default values
			"SalaryPartTime": emp.SalaryPartTime,
			"SalaryAmount":   0.0,
		}

		// If employee has a salary type, get the actual salary amount
		if emp.SalaryPartTime > 0 {
			var salary models.SalaryPartTime
			if err := db.First(&salary, emp.SalaryPartTime).Error; err == nil {
				empData["SalaryAmount"] = salary.Salary
			}
		}

		employeesWithSalary = append(employeesWithSalary, empData)
	}

	c.JSON(http.StatusOK, gin.H{
		"employees": employeesWithSalary,
	})
}

// GetWorkShifts returns all work shifts
func (cmm *Common) GetWorkShifts(c *gin.Context) {
	db := global.Mdb

	// Get all work shifts
	var workShifts []models.WorkShifts
	result := db.Find(&workShifts)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch work shifts",
		})
		return
	}

	// Transform to match frontend expected structure
	type WorkShiftResponse struct {
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		ShiftName string `json:"shiftName"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	}

	var response []WorkShiftResponse
	for _, shift := range workShifts {
		shiftName := fmt.Sprintf("Ca %d (%s - %s)", shift.ID, shift.StartTime, shift.EndTime)
		response = append(response, WorkShiftResponse{
			ID:        shift.ID,
			Name:      shift.ShiftName,
			ShiftName: shiftName,
			StartTime: shift.StartTime,
			EndTime:   shift.EndTime,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"shifts": response,
	})
}
