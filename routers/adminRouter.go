package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/controllers"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/middleware"
)

func ConfigAdminRouter(r *gin.Engine) {

	// Get Admin Controller
	AdminController := controllers.GetNewAdmin()

	// public router
	public := r.Group("/v1")
	{
		public.POST("/signup", AdminController.Signup)
		public.POST("/login", AdminController.Login)

	}

	// private router
	private := r.Group("/v1/admin")
	private.Use(middleware.RequireAuth)
	{
		private.GET("/validate", AdminController.Validate)
		private.POST("/limitem", AdminController.LimitEm)
	}

}
