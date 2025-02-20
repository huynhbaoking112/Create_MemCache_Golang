package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/controllers"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/middleware"
)

func ConfigAdminRouter(r *gin.Engine) {

	// Get Admin Controller
	AdminController := controllers.GetNewAdmin()

	// private router
	private := r.Group("/v1/admin")
	private.Use(middleware.RequireAuth)
	private.Use(middleware.AdminCheck)
	{

		// Đăng ký một nhân viên mới
		private.POST("/signup", AdminController.Signup)

		//---------------------------------------------------------------
		private.GET("/validate", AdminController.Validate)
		// Lấy các đơn xin nghỉ
		private.GET("/takeleave", AdminController.GetTakeLeave)
		// Cho phép nghỉ
		private.POST("/takeleave", AdminController.AcceptTakeleave)

		//---------------------------------------------------------------

		// Giới hạn thành viên trong một ca trong ngày cụ thể
		private.POST("/limitem", AdminController.LimitEm)

		//---------------------------------------------------------------

		// Tạo một lỗi mới
		private.POST("/createnewerror", AdminController.SetNewError)
		// Tạo một lỗi mới dành cho dành cho employee
		private.POST("/createerrofem", AdminController.HanldeErrorEm)

		//---------------------------------------------------------------

		// Tạo bonus cho employee
		private.POST("/createbonus", AdminController.CreateBonus)

		//---------------------------------------------------------------

		// Thanh toán cho employeeemployee
		private.POST("/createpayment", AdminController.PaymentForEm)
		// Tạo SalaryPartTime
		private.POST("/createsalary", AdminController.CreateSalary)
		// Đổi salary cho user
		private.POST("/changesalary", AdminController.ChangeSalary)

		// Cấm User
		private.POST("/activeuser/:id", AdminController.ChangeActive)

	}

}
