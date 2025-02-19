package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/controllers"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/middleware"
)

func ConfigCommonRouter(r *gin.Engine) {

	// Get Common Controller
	CommonController := controllers.GetCommon()

	// public router
	// public := r.Group("/v1/admin")
	// {
	// 	public.POST("/signup", UserController.Signup)
	// 	public.POST("/login", UserController.Login)

	// }

	// private router
	private := r.Group("/v1/common")
	private.Use(middleware.RequireAuth)
	{
		private.GET("/getattendance/:id", CommonController.GetAttendance)
	}

}
