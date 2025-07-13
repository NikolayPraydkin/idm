package main

import (
	"context"
	"fmt"
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/web"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	var server = build()
	// Запускаем сервер в отдельной горутине
	go func() {
		var err = server.App.Listen(":8080")
		if err != nil {
			panic(fmt.Sprintf("http server error: %s", err))
		}
	}()

	// Создаем группу для ожидания сигнала завершения работы сервера
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	// Запускаем gracefulShutdown в отдельной горутине
	go gracefulShutdown(server, wg)
	// Ожидаем сигнал от горутины gracefulShutdown, что сервер завершил работу
	wg.Wait()
	fmt.Println("Graceful shutdown complete.")
}

// Функция "элегантного" завершения работы сервера по сигналу от операционной системы
func gracefulShutdown(server *web.Server, wg *sync.WaitGroup) {
	// Уведомить основную горутину о завершении работы
	defer wg.Done()
	// Создаём контекст, который слушает сигналы прерывания от операционной системы
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()
	// Слушаем сигнал прерывания от операционной системы
	<-ctx.Done()
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")
	// Контекст используется для информирования веб-сервера о том,
	// что у него есть 5 секунд на выполнение запроса, который он обрабатывает в данный момент
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.App.ShutdownWithContext(ctx); err != nil {
		fmt.Printf("Server forced to shutdown with error: %v\n", err)
	}
	fmt.Println("Server exiting")
}

func build() *web.Server {
	var cfg = common.GetConfig(".env")
	var server = web.NewServer()
	var dataBase = database.ConnectDbWithCfg(cfg)

	var employeeRepo = employee.NewEmployeeRepository(dataBase)
	var employeeService = employee.NewService(employeeRepo)
	var employeeController = employee.NewController(server, employeeService)
	employeeController.RegisterRoutes()

	var infoController = info.NewController(server, cfg)
	infoController.RegisterRoutes()

	return server
}
