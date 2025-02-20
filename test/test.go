package main

// import "fmt"
// func (*Admin) PaymentForEm(c *gin.Context) {
// 	// lấy db
// 	db := global.Mdb

// 	// Định nghĩa struct để parse JSON từ request body
// 	var req struct {
// 		EmployeeID   int    `json:"EmployeeID"`
// 		Date         string `json:"Date"`
// 		Time         string `json:"Time"`
// 		Evidence     string `json:"Evidence"`
// 		AttendanceID []int  `json:"attendance_id"`
// 		Bonus        []int  `json:"bonus"`
// 		Error        []int  `json:"error"`
// 	}

// 	// Parse JSON từ request
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Bắt đầu transaction để đảm bảo tính toàn vẹn dữ liệu
// 	tx := db.Begin()

// 	// 1️⃣ Tạo bản ghi mới trong bảng Payment
// 	payment := models.Payment{
// 		EmployeeID: req.EmployeeID,
// 		Date:       req.Date,
// 		Time:       req.Time,
// 		Evidence:   req.Evidence,
// 	}

// 	if err := tx.Create(&payment).Error; err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
// 		return
// 	}

// 	// 2️⃣ Dùng Goroutines + WaitGroup để cập nhật song song
// 	var wg sync.WaitGroup
// 	errChan := make(chan error, 3) // Kênh để nhận lỗi từ Goroutines

// 	wg.Add(3) // 3 công việc chạy song song

// 	// 🏃‍♂️ Cập nhật trạng thái is_payment trong bảng Bonus
// 	go func() {
// 		defer wg.Done()
// 		if len(req.Bonus) > 0 {
// 			if err := tx.Model(&models.Bonus{}).
// 				Where("id IN (?)", req.Bonus).
// 				Update("is_payment", "OK").Error; err != nil {
// 				errChan <- err
// 			}
// 		}
// 	}()

// 	// 🏃‍♂️ Cập nhật trạng thái is_payment trong bảng Error
// 	go func() {
// 		defer wg.Done()
// 		if len(req.Error) > 0 {
// 			if err := tx.Model(&models.Error{}).
// 				Where("id IN (?)", req.Error).
// 				Update("is_payment", "OK").Error; err != nil {
// 				errChan <- err
// 			}
// 		}
// 	}()

// 	// 🏃‍♂️ Tạo các bản ghi trong Payment_Infor
// 	go func() {
// 		defer wg.Done()
// 		var paymentInfoRecords []models.Payment_Infor
// 		for _, attendanceID := range req.AttendanceID {
// 			paymentInfoRecords = append(paymentInfoRecords, models.Payment_Infor{
// 				Id_payment:   int(payment.ID),
// 				AttendanceID: attendanceID,
// 			})
// 		}
// 		for _, bonusID := range req.Bonus {
// 			paymentInfoRecords = append(paymentInfoRecords, models.Payment_Infor{
// 				Id_payment: int(payment.ID),
// 				Bonus:      bonusID,
// 			})
// 		}
// 		for _, errorID := range req.Error {
// 			paymentInfoRecords = append(paymentInfoRecords, models.Payment_Infor{
// 				Id_payment: int(payment.ID),
// 				Error:      errorID,
// 			})
// 		}

// 		if len(paymentInfoRecords) > 0 {
// 			if err := tx.Create(&paymentInfoRecords).Error; err != nil {
// 				errChan <- err
// 			}
// 		}
// 	}()

// 	// Đợi tất cả Goroutines chạy xong
// 	wg.Wait()
// 	close(errChan) // Đóng kênh sau khi tất cả Goroutines hoàn thành

// 	// Kiểm tra nếu có lỗi từ Goroutines
// 	for err := range errChan {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Commit transaction nếu mọi thứ đều thành công
// 	tx.Commit()

// 	c.JSON(http.StatusOK, gin.H{
// 		"message":    "Payment processed successfully",
// 		"payment_id": payment.ID,
// 	})
// }
