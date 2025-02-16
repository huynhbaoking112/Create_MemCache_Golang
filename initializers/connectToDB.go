package initializers

import (
	"fmt"
	"os"

	"github.com/huynhbaoking112/Create_MemCache_Golang.git/global"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDB() {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details

	mysql_port := os.Getenv("MYSQL_PORT")
	mysql_ip := os.Getenv("MYSQL_IP")
	mysql_user := os.Getenv("MYSQL_USER")
	mysql_password := os.Getenv("MYSQL_PASSWORD")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/employee?charset=utf8mb4&parseTime=True&loc=Local", mysql_user, mysql_password, mysql_ip, mysql_port)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("ConnectDB not success")
	}

	global.Mdb = db

	//Migrate DB
	MigrateAllTable()
}

func MigrateAllTable() {
	global.Mdb.AutoMigrate(&models.User{})
}
