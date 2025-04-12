package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/controllers"
)

func ConfigUserRouter(r *gin.Engine) {

	// Get User Controller
	UserController := controllers.GetNewUser()

	// Tạo nhóm v1
	v1 := r.Group("/v1")
	{
		// private router
		private := v1.Group("/user")
		// private.Use(middleware.RequireAuth)
		{
			// Hồ sơ cá nhân
			private.GET("/profile/:userId", UserController.GetProfile)
			private.PUT("/profile", UserController.UpdateProfile)
			private.PUT("/avatar", UserController.UpdateAvatar)

			// Đăng ký ca
			private.POST("/registration", UserController.RegisShift)
			private.GET("/registrations/:userId", UserController.GetUserRegistrations)
			private.DELETE("/registration/:id", UserController.CancelRegistration)

			//--------------------------------------------------------

			// Điểm danh vào
			private.POST("/checkin", UserController.Checkin)
			// Điểm danh ra
			private.POST("/checkout", UserController.Checkout)
			// Lấy thông tin điểm danh
			private.GET("/attendance/:userId", UserController.GetAttendance)

			//--------------------------------------------------------
			// Xin nghỉ
			private.POST("/takeleave", UserController.TakeLeave)
			// Lấy danh sách yêu cầu nghỉ phép
			private.GET("/leaves/:userId", UserController.GetUserLeaveRequests)

			// Thêm các route mới cho thưởng và lỗi
			private.GET("/bonuses/:userId", controllers.GetUserBonuses)
			private.GET("/errors/:userId", controllers.GetUserErrors)

			// Routes cho thanh toán lương
			private.GET("/payments/:userId", controllers.GetUserPaymentHistory)
			private.GET("/payment/:paymentId", controllers.GetUserPaymentDetail)
			private.GET("/payment-items/:paymentId", controllers.GetUserPaymentItems)
		}

		// Thêm route lấy danh sách loại lỗi vào nhóm v1
		v1.GET("/error-types", controllers.GetErrorTypes)
	}
}
