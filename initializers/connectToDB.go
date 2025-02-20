package initializers

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/huynhbaoking112/Create_MemCache_Golang.git/global"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDB() {

	mysql_port := os.Getenv("MYSQL_PORT")
	mysql_ip := os.Getenv("MYSQL_IP")
	mysql_user := os.Getenv("MYSQL_USER")
	mysql_password := os.Getenv("MYSQL_PASSWORD")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/employee?charset=utf8mb4&parseTime=True&loc=Local", mysql_user, mysql_password, mysql_ip, mysql_port)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("ConnectDB not success")
	}

	// Gán biến tổng thểthể
	global.Mdb = db

	// Connection pool
	HandleConnectionPool()

	//Migrate DB
	MigrateAllTable()
}

func HandleConnectionPool() {
	// Lấy `*sql.DB` để cấu hình Connection Pool
	sqlDB, err := global.Mdb.DB()
	if err != nil {
		panic("❌ Failed to get sql.DB")
	}

	// Lấy giá trị từ `.env`
	maxOpenConns, err := strconv.Atoi(os.Getenv("MaxOpenConns"))
	if err != nil {
		maxOpenConns = 100 // Giá trị mặc định nếu parse lỗi
	}

	maxIdleConns, err := strconv.Atoi(os.Getenv("MaxIdleConns"))
	if err != nil {
		maxIdleConns = 10
	}

	connMaxLifetime, err := strconv.Atoi(os.Getenv("ConnMaxLifetime"))
	if err != nil {
		connMaxLifetime = 1 // Giá trị mặc định là 1 giờ
	}

	connMaxIdleTime, err := strconv.Atoi(os.Getenv("ConnMaxIdleTime"))
	if err != nil {
		connMaxIdleTime = 10 // Giá trị mặc định là 10 giờ
	}

	// 🔹 Connection Pool Config (dùng giá trị từ .env)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Duration(connMaxIdleTime) * time.Hour)
}

func MigrateAllTable() {
	global.Mdb.AutoMigrate(&models.Employee{},
		&models.Payment_Infor{},
		&models.Role{},
		&models.SalaryPartTime{},
		&models.WorkShifts{},
		&models.Attendance{},
		&models.Registration{},
		&models.Error{},
		&models.ErrorName{},
		&models.Bonus{},
		&models.Payment{},
		&models.LimitEmployee{},
		&models.Notification{},
		&models.Group{},
		&models.GroupEM{},
		&models.TakeLeave{})
}
