package main

import (
	"context"
	"crypto/tls"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"idm/docs"
	"idm/inner/common"
	"idm/inner/server"
	"idm/inner/web"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// @title IDM API documentation
// @BasePath /api/v1/
func main() {
	var srv, db = server.Build()
	var cfg = common.GetConfig(".env")
	// Переопределяем версию приложения, которая будет отображаться в swagger UI.
	// Пакет docs и структура SwaggerInfo в нём появятся поле генерации документации (см. далее).
	docs.SwaggerInfo.Version = cfg.AppVersion
	var logger = common.NewLogger(cfg)

	// загружаем сертификаты
	cer, err := tls.LoadX509KeyPair(cfg.SslSert, cfg.SslKey)
	if err != nil {
		logger.Panic("failed certificate loading: %s", zap.Error(err))
	}
	// создаём конфигурацию TLS сервера
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
	// создаём слушателя https соединения
	ln, err := tls.Listen("tcp", ":8080", tlsConfig)
	if err != nil {
		logger.Panic("failed TLS listener creating: %s", zap.Error(err))
	}
	// Запускаем сервер в отдельной горутине
	go func() {
		var err = srv.App.Listener(ln)
		if err != nil {
			logger.Panic("http srv error: %s", zap.Error(err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(2)

	go gracefulShutdown(srv, ctx, &wg, logger)
	go closeDb(db, ctx, &wg, logger)

	wg.Wait()
	logger.Sugar().Info("Graceful shutdown complete.")
}

// Функция "элегантного" завершения работы сервера по сигналу от операционной системы
func gracefulShutdown(server *web.Server, ctx context.Context, wg *sync.WaitGroup, logger *common.Logger) {
	// Уведомить основную горутину о завершении работы
	defer wg.Done()
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

func closeDb(db *sqlx.DB, ctx context.Context, wg *sync.WaitGroup, logger *common.Logger) {
	// Уведомить основную горутину о завершении работы
	defer wg.Done()
	// Слушаем сигнал прерывания от операционной системы
	<-ctx.Done()
	err := db.Close()
	if err != nil {
		logger.Sugar().Error("Database close error:", zap.Error(err))
		return
	}
	logger.Sugar().Info("Db closed.")
}

func RegisterMiddleware(app *fiber.App) {
	app.Use(recover.New())
	app.Use(requestid.New())
}
