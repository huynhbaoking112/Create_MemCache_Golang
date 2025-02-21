package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/controllers"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/middleware"
)

func ConfigUserRouter(r *gin.Engine) {

	// Get User Controller
	UserController := controllers.GetNewUser()

	// private router
	private := r.Group("/v1/user")
	private.Use(middleware.RequireAuth)
	{
		// Đăng ký ca
		private.POST("/registration", UserController.RegisShift)

		//--------------------------------------------------------

		// Điểm danh vào
		private.POST("/checkin", UserController.Checkin)
		// Điểm danh ra
		private.POST("/checkout", UserController.Checkout)

		//--------------------------------------------------------
		// Xin nghỉnghỉ
		private.POST("/takeleave", UserController.TakeLeave)
	}

}
