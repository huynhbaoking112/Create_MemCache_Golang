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
	public := r.Group("/v1/common")
	{
		public.POST("/login", CommonController.Login)
	}

	// private router
	private := r.Group("/v1/common")
	private.Use(middleware.RequireAuth)
	{

		// lấy điểm danh
		private.GET("/getattendance/:id", CommonController.GetAttendance)

		// lấy lỗi
		private.GET("/geterrorofem/:id", CommonController.GetErrorOfEm)

		// bonus
		private.GET("/getbonus/:id", CommonController.GetBonusOfEm)

		// check nghỉ không phép
	}

}
