package main

import (
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

	// Admin Router
	routers.ConfigAdminRouter(r)

	// User Router
	routers.ConfigUserRouter(r)

	// Common Router
	routers.ConfigCommonRouter(r)

	r.Run()
}
