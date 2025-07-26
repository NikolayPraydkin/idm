package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"
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
	var cfg = common.GetConfig(".env")
	var logger = common.NewLogger(cfg)
	// Запускаем сервер в отдельной горутине
	go func() {
		var err = server.App.Listen(":8080")
		if err != nil {
			logger.Panic("http server error: %s", zap.Error(err))
		}
	}()

	// Создаем группу для ожидания сигнала завершения работы сервера
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	// Запускаем gracefulShutdown в отдельной горутине
	go gracefulShutdown(server, wg, logger)
	// Ожидаем сигнал от горутины gracefulShutdown, что сервер завершил работу
	wg.Wait()
	logger.Sugar().Info("Graceful shutdown complete.")
}

// Функция "элегантного" завершения работы сервера по сигналу от операционной системы
func gracefulShutdown(server *web.Server, wg *sync.WaitGroup, logger *common.Logger) {
	// Уведомить основную горутину о завершении работы
	defer wg.Done()
	// Создаём контекст, который слушает сигналы прерывания от операционной системы
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()
	// Слушаем сигнал прерывания от операционной системы
	<-ctx.Done()
	logger.Sugar().Info("shutting down gracefully, press Ctrl+C again to force")
	// Контекст используется для информирования веб-сервера о том,
	// что у него есть 5 секунд на выполнение запроса, который он обрабатывает в данный момент
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.App.ShutdownWithContext(ctx); err != nil {
		logger.Sugar().Error("Server forced to shutdown with error:", zap.Error(err))
	}
	logger.Sugar().Info("Server exiting")
}

func build() *web.Server {
	var cfg = common.GetConfig(".env")
	var server = web.NewServer()

	RegisterMiddleware(server.App)

	var dataBase = database.ConnectDbWithCfg(cfg)

	var employeeRepo = employee.NewEmployeeRepository(dataBase)
	var employeeService = employee.NewService(employeeRepo)
	//создаём логгер
	var logger = common.NewLogger(cfg)
	// создаём контроллер
	var employeeController = employee.NewController(server, employeeService, logger)
	employeeController.RegisterRoutes()

	var infoController = info.NewController(server, cfg)
	infoController.RegisterRoutes()

	return server
}

func RegisterMiddleware(app *fiber.App) {
	app.Use(recover.New())
	app.Use(requestid.New())
}
