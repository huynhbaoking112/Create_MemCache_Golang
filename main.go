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

	// Config router Admin
	routers.ConfigAdminRouter(r)

	r.Run()
}
