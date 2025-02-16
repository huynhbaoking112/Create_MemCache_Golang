package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/controllers"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/middleware"
)

func ConfigAdminRouter(r *gin.Engine) {

	AdminController := controllers.GetNewAdmin()

	v1 := r.Group("/v1/admin")
	{
		v1.POST("/signup", AdminController.Signup)
		v1.POST("/login", AdminController.Login)
		v1.GET("/validate", middleware.RequireAuth, AdminController.Validate)
	}

}
