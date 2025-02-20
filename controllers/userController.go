package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/global"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/models"
	"gorm.io/gorm"
)

type User struct {
}

func GetNewUser() *User {
	return &User{}
}

func (*User) RegisShift(c *gin.Context) {
	db := global.Mdb

	var body struct {
		Shift int
		Date  string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found in context"})
		return
	}

	userModel, ok := user.(models.Employee)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to parse user data"})
		return
	}

	// 🏆 Sử dụng goroutines để truy vấn song song
	var wg sync.WaitGroup
	var totalRegistrations int64
	var limitEm models.LimitEmployee
	var checkNV models.Registration
	var errCount, errLimit, errCheck error

	// Channel để nhận kết quả lỗi
	errChan := make(chan error, 3)

	// 🔹 Goroutine 1: Đếm số lượng đăng ký
	wg.Add(1)
	go func() {
		defer wg.Done()
		errCount = db.Model(&models.Registration{}).
			Where("date = ? AND shift = ?", body.Date, body.Shift).
			Count(&totalRegistrations).Error
		if errCount != nil {
			errChan <- errCount
		}
	}()

	// 🔹 Goroutine 2: Kiểm tra giới hạn nhân viên
	wg.Add(1)
	go func() {
		defer wg.Done()
		errLimit = db.Where("date = ? AND shift = ?", body.Date, body.Shift).First(&limitEm).Error
		if errLimit != nil && !errors.Is(errLimit, gorm.ErrRecordNotFound) {
			errChan <- errLimit
		}
	}()

	// 🔹 Goroutine 3: Kiểm tra nhân viên đã đăng ký chưa
	wg.Add(1)
	go func() {
		defer wg.Done()
		errCheck = db.Model(&models.Registration{}).
			Where("employee_id = ? AND date = ? AND shift = ?", userModel.ID, body.Date, body.Shift).
			First(&checkNV).Error
		if errCheck != nil && !errors.Is(errCheck, gorm.ErrRecordNotFound) {
			errChan <- errCheck
		}
	}()

	// 🕒 Chờ tất cả goroutines hoàn thành
	wg.Wait()
	close(errChan)

	// Xử lý lỗi nếu có
	for err := range errChan {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// 🔹 Kiểm tra số lượng đăng ký
	maxRegistrations := 6
	if limitEm.Num != 0 {
		maxRegistrations = limitEm.Num
	}

	if int(totalRegistrations) >= maxRegistrations {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Employee in shift is full"})
		return
	}

	// 🔹 Kiểm tra nhân viên đã đăng ký chưa
	if errCheck == nil { // Nếu tìm thấy bản ghi => nhân viên đã đăng ký
		c.JSON(http.StatusBadRequest, gin.H{"message": "Nhân viên đã đăng ký ca này trước đó"})
		return
	}

	// 🔹 Đăng ký ca làm
	registration := models.Registration{
		EmployeeID: int64(userModel.ID),
		Date:       body.Date,
		Shift:      body.Shift,
	}

	if err := db.Create(&registration).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to register shift"})
		return
	}

	// 🏆 Thành công
	c.JSON(http.StatusOK, gin.H{"message": "Shift registered successfully"})
}
func (*User) Checkout(c *gin.Context) {
	// Lấy DB
	db := global.Mdb

	var body struct {
		EmployeeId int
		Date       string
		Time       string
		Shift      int
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

	body.EmployeeId = int(userModel.ID)

	// Kiểm tra xem nhân viên có đăng ký ca này vào ngày đó không
	var registration models.Registration
	err := db.Where("employee_id = ? AND date = ? AND shift = ?", body.EmployeeId, body.Date, body.Shift).First(&registration).Error
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không đăng ký ca này"})
		return
	}

	var attendance models.Attendance
	// Kiểm tra xem có điểm danh trước đó hay chưa
	errA := db.Where("employee_id = ? AND date = ? AND shift = ?", body.EmployeeId, body.Date, body.Shift).First(&attendance).Error

	if errA != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn đã chưa checkin ca này trước đó"})
		return
	}
	// Kiểm tra xem có checkout trước đó hay chưa
	checkoutTime, _ := time.Parse("15:04:05", attendance.CheckOut)

	checkinTime, _ := time.Parse("15:04:05", attendance.CheckIn)

	if checkoutTime.Sub(checkinTime).Minutes() > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn đã checkout ca này trước đó"})
		return
	}

	// // Kiểm tra về sớmsớm
	var workShift models.WorkShifts
	errShift := db.Where("id = ?", body.Shift).First(&workShift).Error
	if errShift != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy thông tin ca làm việc"})
		return
	}

	// // Chuyển đổi thời gian sang dạng time.time để so sánh
	checkOutTime, err := time.Parse("15:04:05", body.Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Định dạng thời gian check-in không hợp lệ"})
		return
	}
	endTime, err := time.Parse("15:04:05", workShift.EndTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy thời gian bắt đầu ca"})
		return
	}

	// // Tính số phút về sớmớm
	lateMinutes := int(endTime.Sub(checkOutTime).Minutes())

	evi := fmt.Sprintf("Check in at %s", body.Time)
	// // Nếu về sớm  quá 1 tiếng, lưu lỗi vào bảng Violation
	if lateMinutes >= 60 {
		violation := models.Error{
			EmployeeID: body.EmployeeId,
			Date:       body.Date,
			Time:       body.Time,
			TypeError:  3,
			IsPayment:  "NO",
			Evidence:   evi,
		}
		db.Create(&violation)
	} else if
	// quá 10 phút
	lateMinutes >= 10 {
		violation := models.Error{
			EmployeeID: body.EmployeeId,
			Date:       body.Date,
			Time:       body.Time,
			TypeError:  2,
			IsPayment:  "NO",
			Evidence:   evi,
		}
		db.Create(&violation)
	} else if lateMinutes > 5 {
		// quá 5 phút
		violation := models.Error{
			EmployeeID: body.EmployeeId,
			Date:       body.Date,
			Time:       body.Time,
			TypeError:  1,
			IsPayment:  "NO",
			Evidence:   evi,
		}
		db.Create(&violation)
	}

	// Cho check out
	result := db.Model(&models.Attendance{}).Where("employee_id = ? AND date = ? AND shift = ?", body.EmployeeId, body.Date, body.Shift).Update("check_out", body.Time)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi checkout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Checkout thành công"})

}

func (*User) Checkin(c *gin.Context) {
	// Lấy DB
	db := global.Mdb

	var body struct {
		EmployeeId int
		Date       string
		Time       string
		Shift      int
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

	body.EmployeeId = int(userModel.ID)

	// Kiểm tra xem nhân viên có đăng ký ca này vào ngày đó không
	var registration models.Registration
	err := db.Where("employee_id = ? AND date = ? AND shift = ?", body.EmployeeId, body.Date, body.Shift).First(&registration).Error
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không đăng ký ca này"})
		return
	}

	var attendance models.Attendance
	// Kiểm tra xem có điểm danh trước đó hay chưa
	errA := db.Where("employee_id = ? AND date = ? AND shift = ?", body.EmployeeId, body.Date, body.Shift).First(&attendance).Error

	if errA == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn đã checkin ca này trước đó"})
		return
	}

	// Kiểm tra đi trễ
	var workShift models.WorkShifts
	errShift := db.Where("id = ?", body.Shift).First(&workShift).Error
	if errShift != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy thông tin ca làm việc"})
		return
	}

	// Chuyển đổi thời gian sang dạng time.time để so sánh
	checkInTime, err := time.Parse("15:04:05", body.Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Định dạng thời gian check-in không hợp lệ"})
		return
	}
	startTime, err := time.Parse("15:04:05", workShift.StartTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy thời gian bắt đầu ca"})
		return
	}

	// Tính số phút đi trễ
	lateMinutes := int(checkInTime.Sub(startTime).Minutes())

	evi := fmt.Sprintf("Check in at %s", body.Time)
	// Nếu đi trễ quá 5 phút, lưu lỗi vào bảng Violation
	if lateMinutes >= 60 {
		violation := models.Error{
			EmployeeID: body.EmployeeId,
			Date:       body.Date,
			Time:       body.Time,
			TypeError:  3,
			IsPayment:  "NO",
			Evidence:   evi,
		}
		db.Create(&violation)
	} else if
	// quá 10 phút
	lateMinutes >= 10 {
		violation := models.Error{
			EmployeeID: body.EmployeeId,
			Date:       body.Date,
			Time:       body.Time,
			TypeError:  2,
			IsPayment:  "NO",
			Evidence:   evi,
		}
		db.Create(&violation)
	} else if lateMinutes > 5 {
		// quá 5 phút
		violation := models.Error{
			EmployeeID: body.EmployeeId,
			Date:       body.Date,
			Time:       body.Time,
			TypeError:  1,
			IsPayment:  "NO",
			Evidence:   evi,
		}
		db.Create(&violation)
	}

	// Cho check in
	result := db.Create(&models.Attendance{EmployeeID: body.EmployeeId, Date: body.Date, CheckIn: body.Time, Shift: body.Shift})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi checkin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Checkin thành công"})

}

func (*User) TakeLeave(c *gin.Context) {

	// Lấy DB
	db := global.Mdb

	var body struct {
		EMID     int
		Date     string
		Shift    int
		DES      string
		Evidence string
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

	body.EMID = int(userModel.ID)

	// Tìm trong bảng đăng ký ca làm
	var count int64
	err := db.Model(&models.Registration{}).
		Where("employee_id = ? AND date = ? AND shift = ?", body.EMID, body.Date, body.Shift).
		Count(&count).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn cơ sở dữ liệu"})
		return
	}

	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bạn chưa đăng ký ca này"})
		return
	}

	result := db.Create(&models.TakeLeave{EMID: body.EMID, Date: body.Date, Shift: body.Shift, DES: body.DES, Evidience: body.Evidence})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create takeleave",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Create takeleave success",
	})

}
