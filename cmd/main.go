package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/web"
)

func main() {
	// создаём подключение к базе данных
	var database = database.ConnectDb()
	// закрываем соединение с базой данных после выхода из функции main
	defer func() {
		if err := database.Close(); err != nil {
			fmt.Printf("error closing db: %v", err)
		}
	}()
	var server = build(database)
	var err = server.App.Listen(":8080")
	if err != nil {
		panic(fmt.Sprintf("http server error: %s", err))
	}
}

// buil функция, конструирующая наш веб-сервер
func build(database *sqlx.DB) *web.Server {
	// читаем конфиги
	var cfg = common.GetConfig(".env")
	// создаём веб-сервер
	var server = web.NewServer()
	// создаём репозиторий
	var employeeRepo = employee.NewEmployeeRepository(database)
	// создаём сервис
	var employeeService = employee.NewService(employeeRepo)
	// создаём контроллер
	var employeeController = employee.NewController(server, employeeService)
	employeeController.RegisterRoutes()
	// ещё один контроллер
	var infoController = info.NewController(server, cfg)
	infoController.RegisterRoutes()
	return server
}
