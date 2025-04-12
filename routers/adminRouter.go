package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/controllers"
)

func ConfigAdminRouter(r *gin.Engine) {

	// Get Admin Controller
	AdminController := controllers.GetNewAdmin()

	// private router
	private := r.Group("/v1/admin")
	// private.Use(middleware.RequireAuth)
	// private.Use(middleware.AdminCheck)
	{

		// Đăng ký một nhân viên mới
		private.POST("/signup", AdminController.Signup)

		//---------------------------------------------------------------
		private.GET("/validate", AdminController.Validate)
		// Lấy các đơn xin nghỉ
		private.GET("/takeleave", AdminController.GetTakeLeave)
		// Cho phép nghỉ
		private.POST("/takeleave", AdminController.AcceptTakeleave)
		// Từ chối nghỉ
		private.POST("/rejectleave", AdminController.RejectTakeleave)

		//---------------------------------------------------------------

		// Giới hạn thành viên trong một ca trong ngày cụ thể
		private.POST("/limitem", AdminController.LimitEm)
		// Lấy danh sách giới hạn ca làm việc
		private.GET("/shiftlimits", AdminController.GetShiftLimits)
		// Xóa giới hạn ca làm việc
		private.DELETE("/shiftlimit/:id", AdminController.DeleteShiftLimit)

		//---------------------------------------------------------------

		// Tạo một lỗi mới
		private.POST("/createnewerror", AdminController.SetNewError)
		// Lấy danh sách loại lỗi
		private.GET("/error-types", AdminController.GetErrorTypes)
		// Cập nhật loại lỗi
		private.PUT("/error-type/:id", AdminController.UpdateErrorType)
		// Xóa loại lỗi
		private.DELETE("/error-type/:id", AdminController.DeleteErrorType)
		// Tạo một lỗi mới dành cho dành cho employee
		private.POST("/createerrofem", AdminController.HanldeErrorEm)
		// Lấy danh sách lỗi đã ghi nhận
		private.GET("/errors", AdminController.GetEmployeeErrors)

		//---------------------------------------------------------------

		// Tạo bonus cho employee
		private.POST("/createbonus", AdminController.CreateBonus)
		// Lấy danh sách tất cả bonuses
		private.GET("/bonuses", AdminController.GetBonuses)
		// Cập nhật thông tin bonus
		private.PUT("/bonus/:id", AdminController.UpdateBonus)
		// Xóa bonus
		private.DELETE("/bonus/:id", AdminController.DeleteBonus)

		//---------------------------------------------------------------

		// Thanh toán cho employeeemployee
		private.POST("/createpayment", AdminController.PaymentForEm)
		private.GET("/payments", AdminController.GetPayments)
		private.GET("/payments/:id", AdminController.GetPaymentDetails)
		// Get unpaid attendance records for an employee
		private.GET("/unpaid-attendance/:id", AdminController.GetUnpaidAttendance)
		// Get unpaid bonuses for an employee
		private.GET("/unpaid-bonuses/:id", AdminController.GetUnpaidBonuses)
		// Get unpaid errors for an employee
		private.GET("/unpaid-errors/:id", AdminController.GetUnpaidErrors)
		// Get payment information for an employee
		private.GET("/payment-info/:id", AdminController.GetPaymentInfo)
		// Tạo SalaryPartTime
		private.POST("/createsalary", AdminController.CreateSalary)
		// Đổi salary cho user
		private.POST("/changesalary", AdminController.ChangeSalary)

		// Xóa salary
		private.DELETE("/deletesalary/:id", AdminController.DeleteSalary)

		// Cấm User
		private.POST("/activeuser/:id", AdminController.ChangeActive)

		// Cập nhật thông tin nhân viên
		private.POST("/update-employee/:id", AdminController.UpdateEmployee)

		//---------------------------------------------------------------
		// Quản lý đăng ký ca làm việc
		private.GET("/registrations", AdminController.GetRegistrations)
		private.GET("/registrations/:id", AdminController.GetRegistrationByID)
		private.GET("/registrations/employee/:id", AdminController.GetRegistrationsByEmployee)
		private.GET("/registrations/date/:date", AdminController.GetRegistrationsByDate)
		private.GET("/registrations/shift/:shift", AdminController.GetRegistrationsByShift)
		private.POST("/registration", AdminController.CreateRegistration)
		private.PUT("/registration/:id", AdminController.UpdateRegistration)
		private.DELETE("/registration/:id", AdminController.DeleteRegistration)

		// Lấy danh sách nhân viên đăng ký ca nhưng không đi làm
		//---------------------------------------------------------------

	}

}
