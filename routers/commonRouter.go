package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/controllers"
)

func ConfigCommonRouter(r *gin.Engine) {

	// Get Common Controller
	CommonController := controllers.GetCommon()

	// public router
	public := r.Group("/v1/common")
	{
		public.POST("/login", CommonController.Login)

		// Lấy danh sách ca làm việc - endpoint công khai
		public.GET("/shifts", CommonController.GetWorkShifts)
	}

	// private router
	private := r.Group("/v1/common")
	// private.Use(middleware.RequireAuth)
	{
		// lấy điểm danh
		private.GET("/getattendance/:id", CommonController.GetAttendance)

		// lấy lỗi
		private.GET("/geterrorofem/:id", CommonController.GetErrorOfEm)

		// bonus
		private.GET("/getbonus/:id", CommonController.GetBonusOfEm)

		// get all salary types
		private.GET("/salaries", CommonController.GetSalaries)

		// get a specific employee with salary info
		private.GET("/employee/:id", CommonController.GetEmployee)

		// get all employees
		private.GET("/employees", CommonController.GetEmployees)

		// get all employees with salary information
		private.GET("/employees-with-salary", CommonController.GetEmployeesWithSalary)

		// get all work shifts
		private.GET("/workshifts", CommonController.GetWorkShifts)

		// check nghỉ không phép
	}
}
