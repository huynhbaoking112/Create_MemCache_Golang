package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
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

	// Parse request body
	var body struct {
		UserId int    `json:"UserId"`
		Date   string `json:"Date"`
		Shift  int    `json:"Shift"`
	}

	// Log raw request data trước khi binding
	bodyBytes, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không thể đọc dữ liệu request",
			"error":   err.Error(),
		})
		return
	}

	// Log ra body raw
	fmt.Printf("Raw request body: %s\n", string(bodyBytes))

	// Tạo lại reader cho binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Bind body
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Dữ liệu không hợp lệ, vui lòng kiểm tra lại định dạng JSON",
			"error":   err.Error(),
		})
		return
	}

	// In ra thông tin body đã parse
	fmt.Printf("Parsed body: UserId=%d, Date=%s, Shift=%d\n", body.UserId, body.Date, body.Shift)

	// Kiểm tra dữ liệu hợp lệ
	if body.UserId == 0 || body.Date == "" || body.Shift < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Dữ liệu không hợp lệ, vui lòng kiểm tra lại",
			"data": gin.H{
				"received": gin.H{
					"UserId": body.UserId,
					"Date":   body.Date,
					"Shift":  body.Shift,
				},
				"required": "UserId > 0, Date không được rỗng, Shift > 0",
			},
		})
		return
	}

	// Kiểm tra nhân viên tồn tại
	var employee models.Employee
	if err := db.First(&employee, body.UserId).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không tìm thấy thông tin nhân viên",
			"error":   err.Error(),
		})
		return
	}

	// Kiểm tra ca làm việc tồn tại
	var shift models.WorkShifts
	if err := db.First(&shift, body.Shift).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không tìm thấy ca làm việc",
			"error":   err.Error(),
		})
		return
	}

	// Validate date format
	regDate, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Ngày không hợp lệ, sử dụng định dạng YYYY-MM-DD",
			"error":   err.Error(),
		})
		return
	}

	// Không cho phép đăng ký ca trong quá khứ
	currentDate := time.Now()
	if regDate.Before(time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không thể đăng ký ca làm việc cho ngày trong quá khứ",
			"data": gin.H{
				"requestDate": body.Date,
				"currentDate": currentDate.Format("2006-01-02"),
			},
		})
		return
	}

	// Kiểm tra nhân viên đã đăng ký ca này chưa
	var existingRegistration models.Registration
	err = db.Where("employee_id = ? AND date = ? AND shift = ?", body.UserId, body.Date, body.Shift).
		First(&existingRegistration).Error
	if err == nil {
		// Đã tìm thấy đăng ký => nhân viên đã đăng ký
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Bạn đã đăng ký ca này trước đó",
			"data": gin.H{
				"registration": gin.H{
					"id":         existingRegistration.ID,
					"employeeId": existingRegistration.EmployeeID,
					"date":       existingRegistration.Date,
					"shift":      existingRegistration.Shift,
				},
			},
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Lỗi khác, không phải "không tìm thấy"
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Lỗi khi kiểm tra đăng ký hiện có",
			"error":   err.Error(),
		})
		return
	}

	// Kiểm tra giới hạn đăng ký ca
	var limitEmployee models.LimitEmployee
	var shiftLimit int = 6 // Mặc định giới hạn là 6

	err = db.Where("date = ? AND shift = ?", body.Date, body.Shift).First(&limitEmployee).Error
	if err == nil {
		// Có giới hạn cụ thể trong cơ sở dữ liệu
		shiftLimit = limitEmployee.Num
	}

	// Đếm số lượng đăng ký hiện tại
	var registeredCount int64
	db.Model(&models.Registration{}).
		Where("date = ? AND shift = ?", body.Date, body.Shift).
		Count(&registeredCount)

	// Kiểm tra có vượt quá giới hạn không
	if int(registeredCount) >= shiftLimit {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("Ca làm việc này đã đạt giới hạn %d nhân viên", shiftLimit),
			"data": gin.H{
				"limit":     shiftLimit,
				"current":   registeredCount,
				"remaining": 0,
			},
		})
		return
	}

	// Tạo đăng ký mới
	registration := models.Registration{
		EmployeeID: int64(body.UserId),
		Date:       body.Date,
		Shift:      body.Shift,
	}

	if err := db.Create(&registration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể tạo đăng ký",
			"error":   err.Error(),
		})
		return
	}

	// Log thông tin đăng ký thành công
	fmt.Printf("Đăng ký thành công: ID=%d, EmployeeID=%d, Date=%s, Shift=%d\n",
		registration.ID, registration.EmployeeID, registration.Date, registration.Shift)

	// Tạo tên ca làm việc
	shiftName := fmt.Sprintf("Ca %d (%s - %s)",
		shift.ID,
		shift.StartTime,
		shift.EndTime)

	// Thành công
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Đăng ký ca làm việc thành công",
		"registration": gin.H{
			"id":        registration.ID,
			"date":      registration.Date,
			"shift":     registration.Shift,
			"shiftName": shiftName,
		},
		"remaining": shiftLimit - int(registeredCount) - 1,
		"limit":     shiftLimit,
	})
}

func (*User) Checkout(c *gin.Context) {
	// Lấy DB
	db := global.Mdb

	var body struct {
		EmployeeId int
		Date       string
		Time       string
		Shift      int
		UserId     int
	}

	// Móc dữ liệu
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}
	// // Lấy giá trị user từ context
	// user, exists := c.Get("user")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{
	// 		"message": "User not found in context",
	// 	})
	// 	return
	// }

	// // Ép kiểu user về models.Employee
	// userModel, ok := user.(models.Employee)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"message": "Failed to parse user data",
	// 	})
	// 	return
	// }

	// body.EmployeeId = int(userModel.ID)

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
	// // Lấy giá trị user từ context
	// user, exists := c.Get("user")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{
	// 		"message": "User not found in context",
	// 	})
	// 	return
	// }

	// // Ép kiểu user về models.Employee
	// userModel, ok := user.(models.Employee)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"message": "Failed to parse user data",
	// 	})
	// 	return
	// }

	// body.EmployeeId = int(userModel.ID)

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
	result := db.Create(&models.Attendance{EmployeeID: uint(body.EmployeeId), Date: body.Date, CheckIn: body.Time, Shift: body.Shift})

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
	// user, exists := c.Get("user")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{
	// 		"message": "User not found in context",
	// 	})
	// 	return
	// }

	// Ép kiểu user về models.Employee
	// userModel, ok := user.(models.Employee)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"message": "Failed to parse user data",
	// 	})
	// 	return
	// }

	// body.EMID = int(userModel.ID)

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

// GetProfile lấy thông tin profile của user
func (*User) GetProfile(c *gin.Context) {
	db := global.Mdb

	// Lấy userId từ param
	userId := c.Param("userId")

	// Lấy thông tin user từ database
	var user models.Employee
	result := db.First(&user, userId)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Không tìm thấy thông tin người dùng",
		})
		return
	}

	// Trả về thông tin user (không bao gồm password)
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    user,
	})
}

// UpdateProfile cập nhật thông tin cá nhân của user
func (*User) UpdateProfile(c *gin.Context) {
	db := global.Mdb

	// Lấy dữ liệu cập nhật từ request body
	var updateData struct {
		UserID      int    `json:"userId"`
		FullName    string `json:"fullName"`
		Gender      string `json:"gender"`
		Phone       string `json:"phone"`
		Address     string `json:"address"`
		DateOfBirth string `json:"dateOfBirth"`
		CardNumber  string `json:"cardNumber"`
		Bank        string `json:"bank"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Dữ liệu không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Kiểm tra ID người dùng có tồn tại
	if updateData.UserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không tìm thấy thông tin người dùng",
		})
		return
	}

	// Cập nhật thông tin
	updates := map[string]interface{}{
		"full_name":     updateData.FullName,
		"gender":        updateData.Gender,
		"phone":         updateData.Phone,
		"address":       updateData.Address,
		"date_of_birth": updateData.DateOfBirth,
		"card_number":   updateData.CardNumber,
		"bank":          updateData.Bank,
	}

	// Chỉ cập nhật các trường không rỗng
	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if strValue, ok := value.(string); ok && strValue != "" {
			filteredUpdates[key] = value
		}
	}

	if len(filteredUpdates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không có thông tin nào được cập nhật",
		})
		return
	}

	// Cập nhật vào database
	result := db.Model(&models.Employee{}).Where("id = ?", updateData.UserID).Updates(filteredUpdates)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể cập nhật thông tin",
			"error":   result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không có thông tin nào được cập nhật",
		})
		return
	}

	// Lấy thông tin user sau khi cập nhật
	var updatedUser models.Employee
	db.First(&updatedUser, updateData.UserID)
	updatedUser.Password = "" // Không trả về password

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cập nhật thông tin thành công",
		"user":    updatedUser,
	})
}

// UpdateAvatar cập nhật ảnh đại diện của user
func (*User) UpdateAvatar(c *gin.Context) {
	db := global.Mdb

	// Lấy URL ảnh từ request body
	var updateData struct {
		UserID int    `json:"userId"`
		Image  string `json:"image"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Dữ liệu không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Kiểm tra ID người dùng có tồn tại
	if updateData.UserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không tìm thấy thông tin người dùng",
		})
		return
	}

	if updateData.Image == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "URL ảnh không được để trống",
		})
		return
	}

	// Cập nhật ảnh đại diện vào database
	result := db.Model(&models.Employee{}).Where("id = ?", updateData.UserID).Update("image", updateData.Image)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể cập nhật ảnh đại diện",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cập nhật ảnh đại diện thành công",
		"image":   updateData.Image,
	})
}

// GetUserRegistrations lấy danh sách đăng ký ca của người dùng
func (*User) GetUserRegistrations(c *gin.Context) {
	db := global.Mdb

	// Lấy userId từ param
	userId := c.Param("userId")

	// Tham số tìm kiếm
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Query cơ sở
	query := db.Model(&models.Registration{}).
		Where("employee_id = ?", userId).
		Order("date DESC, shift ASC")

	// Thêm bộ lọc nếu có
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}

	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	// Lấy danh sách đăng ký
	var registrations []struct {
		models.Registration
		ShiftName string `json:"shiftName"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	}

	err := query.
		Joins("LEFT JOIN work_shifts ON work_shifts.id = registrations.shift").
		Select("registrations.*, work_shifts.shift_name as ShiftName, work_shifts.start_time as StartTime, work_shifts.end_time as EndTime").
		Scan(&registrations).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể lấy danh sách đăng ký",
			"error":   err.Error(),
		})
		return
	}

	// Định dạng tên ca làm việc
	for i := range registrations {
		if registrations[i].ShiftName == "" {
			// Nếu không có tên ca, tạo tên ca dựa trên ID
			registrations[i].ShiftName = fmt.Sprintf("Ca %d", registrations[i].Shift)
		}

		// Nếu có thời gian bắt đầu và kết thúc, thì thêm vào tên ca
		if registrations[i].StartTime != "" && registrations[i].EndTime != "" {
			registrations[i].ShiftName = fmt.Sprintf("Ca %d (%s - %s)",
				registrations[i].Shift,
				registrations[i].StartTime,
				registrations[i].EndTime)
		}
	}

	// Kiểm tra xem có đăng ký nào không
	if len(registrations) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"message":       "Không có đăng ký nào",
			"registrations": []struct{}{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"registrations": registrations,
	})
}

// CancelRegistration hủy đăng ký ca làm việc
func (*User) CancelRegistration(c *gin.Context) {
	db := global.Mdb

	// Lấy ID của đăng ký từ param
	registrationId := c.Param("id")

	// Kiểm tra xem đăng ký có tồn tại không
	var registration models.Registration
	err := db.First(&registration, registrationId).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Không tìm thấy đăng ký ca làm việc",
		})
		return
	}

	// Kiểm tra ngày đăng ký, không cho phép hủy đăng ký cho ngày hiện tại hoặc quá khứ
	currentDate := time.Now().Format("2006-01-02")
	if registration.Date <= currentDate {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Không thể hủy đăng ký cho ngày hiện tại hoặc quá khứ",
		})
		return
	}

	// Xóa đăng ký
	if err := db.Delete(&registration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể hủy đăng ký ca làm việc",
			"error":   err.Error(),
		})
		return
	}

	// Lấy thông tin ca làm việc để thêm vào kết quả
	var shift models.WorkShifts
	db.First(&shift, registration.Shift)

	shiftName := fmt.Sprintf("Ca %d", registration.Shift)
	if shift.ID > 0 {
		shiftName = fmt.Sprintf("Ca %d (%s - %s)",
			shift.ID,
			shift.StartTime,
			shift.EndTime)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Hủy đăng ký ca làm việc thành công",
		"data": gin.H{
			"id":        registration.ID,
			"date":      registration.Date,
			"shift":     registration.Shift,
			"shiftName": shiftName,
		},
	})
}

// GetUserBonuses lấy danh sách thưởng của người dùng
func GetUserBonuses(c *gin.Context) {
	userID := c.Param("userId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var bonuses []models.Bonus
	query := global.Mdb.Where("employee_id = ?", userID)

	if startDate != "" && endDate != "" {
		query = query.Where("date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Find(&bonuses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Không thể lấy danh sách thưởng",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, bonuses)
}

// GetUserErrors lấy danh sách lỗi của người dùng
func GetUserErrors(c *gin.Context) {
	userID := c.Param("userId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// Truy vấn danh sách lỗi
	var errors []models.Error
	query := global.Mdb.Where("employee_id = ?", userID)

	if startDate != "" && endDate != "" {
		query = query.Where("date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Find(&errors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Không thể lấy danh sách lỗi",
			"error":   err.Error(),
		})
		return
	}

	// Lấy thông tin chi tiết cho mỗi lỗi
	type ErrorWithDetails struct {
		models.Error
		ErrorName string  `json:"errorName"`
		Fines     float64 `json:"fines"`
	}

	var result []ErrorWithDetails
	for _, err := range errors {
		var errorName models.ErrorName
		if errFind := global.Mdb.First(&errorName, err.TypeError).Error; errFind != nil {
			// Nếu không tìm thấy thông tin lỗi, vẫn thêm lỗi nhưng không có tên
			result = append(result, ErrorWithDetails{
				Error:     err,
				ErrorName: "Không xác định",
				Fines:     0,
			})
		} else {
			result = append(result, ErrorWithDetails{
				Error:     err,
				ErrorName: errorName.NameError,
				Fines:     errorName.Fines,
			})
		}
	}

	c.JSON(http.StatusOK, result)
}

// GetErrorTypes lấy danh sách loại lỗi
func GetErrorTypes(c *gin.Context) {
	var errorTypes []models.ErrorName

	if err := global.Mdb.Find(&errorTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Không thể lấy danh sách loại lỗi",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, errorTypes)
}

// GetUserPaymentHistory Lấy lịch sử thanh toán của user
func GetUserPaymentHistory(c *gin.Context) {
	userId := c.Param("userId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	// Phân trang
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Kiểm tra user có tồn tại
	var employee models.Employee
	if err := global.Mdb.First(&employee, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Employee not found",
		})
		return
	}

	// Lấy danh sách thanh toán
	var payments []models.Payment
	query := global.Mdb.Where("employee_id = ?", userId)

	// Thêm filter ngày nếu có
	if startDate != "" && endDate != "" {
		query = query.Where("date >= ? AND date <= ?", startDate, endDate)
	}

	// Đếm tổng số kết quả
	var total int64
	query.Model(&models.Payment{}).Count(&total)

	// Lấy danh sách với phân trang
	if err := query.Order("date DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get payment history",
			"error":   err.Error(),
		})
		return
	}

	// Lấy thông tin chi tiết cho mỗi thanh toán
	var paymentResponses []gin.H
	for _, payment := range payments {
		// Đếm số bản ghi trong Payment_Infor cho payment này
		var attendanceCount, bonusCount, errorCount int64
		var paymentInfos []models.Payment_Infor

		global.Mdb.Where("id_payment = ?", payment.ID).Find(&paymentInfos)

		for _, info := range paymentInfos {
			if info.AttendanceID > 0 {
				attendanceCount++
			}
			if info.Bonus > 0 {
				bonusCount++
			}
			if info.Error > 0 {
				errorCount++
			}
		}

		// Tính tổng tiền (nếu cần)
		// Để đơn giản, không tính toán ở đây, sẽ tính trong chi tiết

		paymentResponses = append(paymentResponses, gin.H{
			"id":              payment.ID,
			"employeeId":      payment.EmployeeID,
			"date":            payment.Date,
			"time":            payment.Time,
			"evidence":        payment.Evidence,
			"attendanceCount": attendanceCount,
			"bonusCount":      bonusCount,
			"errorCount":      errorCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"payments": paymentResponses,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetUserPaymentDetail Lấy chi tiết thanh toán
func GetUserPaymentDetail(c *gin.Context) {
	paymentId := c.Param("paymentId")

	// Lấy thông tin thanh toán
	var payment models.Payment
	if err := global.Mdb.First(&payment, paymentId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Payment not found",
		})
		return
	}

	// Lấy thông tin nhân viên
	var employee models.Employee
	if err := global.Mdb.First(&employee, payment.EmployeeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Employee not found",
		})
		return
	}

	// Lấy thông tin lương giờ
	var salaryPartTime models.SalaryPartTime
	if err := global.Mdb.First(&salaryPartTime, employee.SalaryPartTime).Error; err == nil {
		employee.SalaryPartTime = int(salaryPartTime.Salary)
	}

	// Lấy payment info
	var paymentInfos []models.Payment_Infor
	if err := global.Mdb.Where("id_payment = ?", paymentId).Find(&paymentInfos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get payment info",
		})
		return
	}

	// Lấy danh sách attendance, bonus, error từ payment_info
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

	// Lấy chi tiết điểm danh
	var attendances []gin.H
	if len(attendanceIDs) > 0 {
		var attendanceList []models.Attendance
		if err := global.Mdb.Where("id IN ?", attendanceIDs).Find(&attendanceList).Error; err == nil {
			for _, attendance := range attendanceList {
				// Lấy thông tin ca làm
				var shift models.WorkShifts
				global.Mdb.First(&shift, attendance.Shift)

				attendances = append(attendances, gin.H{
					"id":       attendance.ID,
					"date":     attendance.Date,
					"shift":    shift.ShiftName,
					"checkIn":  attendance.CheckIn,
					"checkOut": attendance.CheckOut,
				})
			}
		}
	}

	// Lấy chi tiết thưởng
	var bonuses []gin.H
	if len(bonusIDs) > 0 {
		var bonusList []models.Bonus
		if err := global.Mdb.Where("id IN ?", bonusIDs).Find(&bonusList).Error; err == nil {
			for _, bonus := range bonusList {
				bonuses = append(bonuses, gin.H{
					"id":          bonus.ID,
					"date":        bonus.Date,
					"description": bonus.Description,
					"money":       bonus.Money,
				})
			}
		}
	}

	// Lấy chi tiết lỗi
	var errors []gin.H
	if len(errorIDs) > 0 {
		var errorList []models.Error
		if err := global.Mdb.Where("id IN ?", errorIDs).Find(&errorList).Error; err == nil {
			for _, errItem := range errorList {
				// Lấy tên lỗi
				var errorName models.ErrorName
				global.Mdb.First(&errorName, errItem.TypeError)

				errors = append(errors, gin.H{
					"id":        errItem.ID,
					"date":      errItem.Date,
					"nameError": errorName.NameError,
					"fines":     errorName.Fines,
					"evidence":  errItem.Evidence,
				})
			}
		}
	}

	// Bỏ password trước khi trả về
	employee.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"payment":    payment,
		"employee":   employee,
		"attendance": attendances,
		"bonus":      bonuses,
		"error":      errors,
	})
}

// GetUserPaymentItems lấy các mục chi tiết của một thanh toán
func GetUserPaymentItems(c *gin.Context) {
	paymentID := c.Param("paymentId")

	// Kiểm tra và chuyển đổi paymentID
	var paymentIDInt int
	if _, err := fmt.Sscanf(paymentID, "%d", &paymentIDInt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID thanh toán không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Lấy tất cả các mục trong payment_infor
	var paymentInfos []models.Payment_Infor
	if err := global.Mdb.Where("id_payment = ?", paymentIDInt).
		Find(&paymentInfos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể lấy các mục thanh toán",
			"error":   err.Error(),
		})
		return
	}

	// Trả về danh sách các mục
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"items":   paymentInfos,
		"count":   len(paymentInfos),
	})
}

// GetAttendance lấy thông tin điểm danh của nhân viên
func (*User) GetAttendance(c *gin.Context) {
	db := global.Mdb

	// Lấy userId từ param
	userId := c.Param("userId")

	// Kiểm tra nếu có tham số date thì lấy điểm danh theo ngày đó
	date := c.Query("date")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var attendances []models.Attendance
	var result *gorm.DB

	if date != "" {
		// Lấy điểm danh theo ngày cụ thể
		result = db.Where("employee_id = ? AND date = ?", userId, date).Find(&attendances)
	} else if startDate != "" && endDate != "" {
		// Lấy điểm danh trong khoảng thời gian
		result = db.Where("employee_id = ? AND date BETWEEN ? AND ?", userId, startDate, endDate).Find(&attendances)
	} else {
		// Lấy tất cả điểm danh
		result = db.Where("employee_id = ?", userId).Find(&attendances)
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Lỗi khi lấy dữ liệu điểm danh",
			"error":   result.Error.Error(),
		})
		return
	}

	// Lấy thông tin ca làm việc cho mỗi điểm danh
	type AttendanceResponse struct {
		ID         uint    `json:"id"`
		EmployeeID uint    `json:"employeeId"`
		Date       string  `json:"date"`
		Shift      int     `json:"shift"`
		ShiftName  string  `json:"shiftName"`
		StartTime  string  `json:"startTime"`
		EndTime    string  `json:"endTime"`
		CheckIn    string  `json:"checkIn"`
		CheckOut   string  `json:"checkOut"`
		WorkHours  float64 `json:"workHours"`
		IsLate     bool    `json:"isLate"`
		IsEarlyOut bool    `json:"isEarlyOut"`
	}

	var attendanceResponses []AttendanceResponse

	for _, attendance := range attendances {
		// Lấy thông tin ca làm việc
		var shift models.WorkShifts
		if err := db.Where("id = ?", attendance.Shift).First(&shift).Error; err != nil {
			continue
		}

		// Tính toán giờ làm việc
		var workHours float64 = 0
		var isLate bool = false
		var isEarlyOut bool = false

		if attendance.CheckIn != "" && attendance.CheckOut != "" {
			checkIn, _ := time.Parse("15:04:05", attendance.CheckIn)
			checkOut, _ := time.Parse("15:04:05", attendance.CheckOut)
			startTime, _ := time.Parse("15:04:05", shift.StartTime)
			endTime, _ := time.Parse("15:04:05", shift.EndTime)

			// Kiểm tra đi trễ
			if checkIn.After(startTime) {
				isLate = true
			}

			// Kiểm tra về sớm
			if checkOut.Before(endTime) {
				isEarlyOut = true
			}

			// Tính giờ làm việc
			hours := checkOut.Sub(checkIn).Hours()
			workHours = math.Max(0, hours) // Đảm bảo không âm
		}

		attendanceResponses = append(attendanceResponses, AttendanceResponse{
			ID:         attendance.ID,
			EmployeeID: attendance.EmployeeID,
			Date:       attendance.Date,
			Shift:      attendance.Shift,
			ShiftName:  shift.ShiftName,
			StartTime:  shift.StartTime,
			EndTime:    shift.EndTime,
			CheckIn:    attendance.CheckIn,
			CheckOut:   attendance.CheckOut,
			WorkHours:  workHours,
			IsLate:     isLate,
			IsEarlyOut: isEarlyOut,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"attendance": attendanceResponses,
	})
}

// GetUserLeaveRequests lấy danh sách yêu cầu nghỉ phép của người dùng
func (*User) GetUserLeaveRequests(c *gin.Context) {
	db := global.Mdb

	// Lấy userId từ param
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Thiếu thông tin người dùng",
		})
		return
	}

	// Chuyển đổi userId sang integer
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Mã người dùng không hợp lệ",
			"error":   err.Error(),
		})
		return
	}

	// Chuẩn bị query
	query := db.Table("take_leaves").
		Where("em_id = ?", userIdInt).
		Order("created_at DESC")

	// Kiểm tra nếu có startDate
	startDate := c.Query("startDate")
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}

	// Kiểm tra nếu có endDate
	endDate := c.Query("endDate")
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	// Thực hiện query
	var leaveRequests []struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"createdAt"`
		EMID      int       `json:"employeeId"`
		Date      string    `json:"date"`
		Shift     int       `json:"shift"`
		DES       string    `json:"description"`
		Evidience string    `json:"evidence"`
		IsAgree   string    `json:"status"`
	}

	if err := query.Find(&leaveRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Không thể lấy danh sách yêu cầu nghỉ phép",
			"error":   err.Error(),
		})
		return
	}

	// Lấy thông tin chi tiết về ca làm việc (hỗ trợ hiển thị tên ca)
	for i, leave := range leaveRequests {
		var shift models.WorkShifts
		if err := db.First(&shift, leave.Shift).Error; err == nil {
			// Tạo một phiên bản mới của struct để cập nhật
			leaveWithShift := leave
			leaveWithShift.IsAgree = formatLeaveStatus(leave.IsAgree)
			leaveRequests[i] = leaveWithShift
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"leaves":  leaveRequests,
	})
}

// Hàm hỗ trợ định dạng trạng thái nghỉ phép
func formatLeaveStatus(status string) string {
	switch status {
	case "OK":
		return "Đã duyệt"
	case "NO":
		return "Đã từ chối"
	case "PROCESS":
		return "Đang xử lý"
	default:
		return "Không xác định"
	}
}
