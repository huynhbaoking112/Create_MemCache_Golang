package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/initializers"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/routers"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectDB()
}

func main() {
	r := gin.Default()

	// Configure CORS to allow all origins
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false, // Set to false when AllowOrigins is "*"
		MaxAge:           86400, // 24 hours
	}))

	// Admin Router
	routers.ConfigAdminRouter(r)

	// User Router
	routers.ConfigUserRouter(r)

	// Common Router
	routers.ConfigCommonRouter(r)

	r.Run(":8000")
}
