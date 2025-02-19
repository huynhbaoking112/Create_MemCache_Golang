package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/controllers"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/middleware"
)

func ConfigUserRouter(r *gin.Engine) {

	// Get Admin Controller
	UserController := controllers.GetNewUser()

	// public router
	// public := r.Group("/v1/admin")
	// {
	// 	public.POST("/signup", UserController.Signup)
	// 	public.POST("/login", UserController.Login)

	// }

	// private router
	private := r.Group("/v1/user")
	private.Use(middleware.RequireAuth)
	{
		private.POST("/registration", UserController.RegisShift)
		private.POST("/checkin", UserController.Checkin)
		private.POST("/checkout", UserController.Checkout)
		private.POST("/takeleave", UserController.TakeLeave)
	}

}
