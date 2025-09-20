package server

import (
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/web"

	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"github.com/jmoiron/sqlx"
)

func Build() (*web.Server, *sqlx.DB) {
	var cfg = common.GetConfig(".env")
	//создаём логгер
	var logger = common.NewLogger(cfg)
	var server = web.NewServer()

	server.App.Use("/swagger/*", swagger.HandlerDefault)
	server.App.Use(requestid.New())
	server.App.Use(recover.New())
	server.GroupApi.Use(web.AuthMiddleware(logger))

	var db = database.ConnectDbWithCfg(cfg)

	var employeeRepo = employee.NewEmployeeRepository(db)
	var employeeService = employee.NewService(employeeRepo)

	// создаём контроллер
	var employeeController = employee.NewController(server, employeeService, logger)
	employeeController.RegisterRoutes()

	var infoController = info.NewController(server, cfg)
	infoController.RegisterRoutes()

	return server, db
}
