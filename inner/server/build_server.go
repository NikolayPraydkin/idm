package server

import (
	"github.com/jmoiron/sqlx"
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/web"
)

func Build() (*web.Server, *sqlx.DB) {
	var cfg = common.GetConfig(".env")
	var server = web.NewServer()

	var db = database.ConnectDbWithCfg(cfg)

	var employeeRepo = employee.NewEmployeeRepository(db)
	var employeeService = employee.NewService(employeeRepo)
	//создаём логгер
	var logger = common.NewLogger(cfg)
	// создаём контроллер
	var employeeController = employee.NewController(server, employeeService, logger)
	employeeController.RegisterRoutes()

	var infoController = info.NewController(server, cfg)
	infoController.RegisterRoutes()

	return server, db
}
