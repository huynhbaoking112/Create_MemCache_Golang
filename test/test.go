package test

// import (
// 	"fmt"
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/huynhbaoking112/Create_MemCache_Golang.git/global"
// 	"github.com/huynhbaoking112/Create_MemCache_Golang.git/models"
// 	"gorm.io/gorm"
// )

// // Định nghĩa khung giờ cho các ca làm việc
// var shifts = map[int]struct {
// 	Start string
// 	End   string
// }{
// 	1: {"06:00:00", "12:00:00"},
// 	2: {"12:00:00", "18:00:00"},
// 	3: {"18:00:00", "22:00:00"},
// }

// // Hàm chấm công
// func (u *User) Checkin(c *gin.Context) {
// 	// Lấy DB từ global
// 	db := global.Mdb

// 	var body struct {
// 		EmployeeID int
// 		Date       string
// 		Time       string
// 	}

// 	// Đọc dữ liệu từ request
// 	if err := c.Bind(&body); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
// 		return
// 	}

// 	// Lấy thông tin user từ context
// 	user, exists := c.Get("user")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found in context"})
// 		return
// 	}

// 	// Ép kiểu user về models.Employee
// 	userModel, ok := user.(models.Employee)
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to parse user data"})
// 		return
// 	}

// 	// Gán ID của user vào body
// 	body.EmployeeID = int(userModel.ID)

// 	// Kiểm tra xem nhân viên đã đăng ký ca nào trong ngày chưa
// 	var registrations []Registration
// 	err := db.Where("employee_id = ? AND date = ?", body.EmployeeID, body.Date).Find(&registrations).Error
// 	if err != nil || len(registrations) == 0 {
// 		c.JSON(http.StatusForbidden, gin.H{"message": "Employee has not registered any shifts today"})
// 		return
// 	}

// 	// Chuyển đổi thời gian nhập vào
// 	checkTime, err := time.Parse("15:04:05", body.Time)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid time format"})
// 		return
// 	}

// 	// Tìm ca làm hợp lệ
// 	var selectedShift *Registration
// 	for _, reg := range registrations {
// 		shift, exists := shifts[reg.Shift]
// 		if !exists {
// 			continue
// 		}

// 		startTime, _ := time.Parse("15:04:05", shift.Start)
// 		endTime, _ := time.Parse("15:04:05", shift.End)

// 		// Kiểm tra nếu thời gian nhập vào nằm trong ca này
// 		if checkTime.After(startTime.Add(-5*time.Minute)) && checkTime.Before(endTime.Add(5*time.Minute)) {
// 			selectedShift = &reg
// 			break
// 		}
// 	}

// 	// Nếu không có ca phù hợp, trả về lỗi
// 	if selectedShift == nil {
// 		c.JSON(http.StatusForbidden, gin.H{"message": "Invalid shift time"})
// 		return
// 	}

// 	// Lấy thông tin ca làm việc
// 	shift := shifts[selectedShift.Shift]
// 	startTime, _ := time.Parse("15:04:05", shift.Start)
// 	endTime, _ := time.Parse("15:04:05", shift.End)

// 	// Kiểm tra đi trễ hoặc về sớm
// 	if checkTime.After(startTime.Add(5 * time.Minute)) {
// 		fmt.Println("🚨 Nhân viên đi trễ 🚨")
// 	}

// 	if checkTime.Before(endTime.Add(-5 * time.Minute)) {
// 		fmt.Println("🚨 Nhân viên về sớm 🚨")
// 	}

// 	// Kiểm tra xem đã có check-in trước đó không
// 	var attendance Attendance
// 	result := db.Where("employee_id = ? AND date = ? AND shift = ?", body.EmployeeID, body.Date, selectedShift.Shift).First(&attendance)

// 	if result.Error == gorm.ErrRecordNotFound {
// 		// Nếu chưa có thì tạo mới với Check-in
// 		attendance = Attendance{
// 			EmployeeID: body.EmployeeID,
// 			Date:       body.Date,
// 			CheckIn:    &body.Time,
// 			Shift:      selectedShift.Shift,
// 		}
// 		if err := db.Create(&attendance).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to record check-in"})
// 			return
// 		}
// 		c.JSON(http.StatusOK, gin.H{"message": "Check-in recorded successfully"})
// 		return
// 	}

// 	// Nếu đã có check-in nhưng chưa có check-out, thì cập nhật Check-out
// 	if attendance.CheckOut == nil {
// 		attendance.CheckOut = &body.Time
// 		if err := db.Save(&attendance).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to record check-out"})
// 			return
// 		}
// 		c.JSON(http.StatusOK, gin.H{"message": "Check-out recorded successfully"})
// 		return
// 	}

// 	// Nếu đã có cả check-in và check-out, không cho chấm công lại
// 	c.JSON(http.StatusBadRequest, gin.H{"message": "Attendance already completed for this shift"})
// }
