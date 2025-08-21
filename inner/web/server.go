package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "idm/docs"
)

// структуа веб-сервера
type Server struct {
	App           *fiber.App
	GroupApiV1    fiber.Router
	GroupInternal fiber.Router
}

// функция-конструктор
func NewServer() *Server {

	// создаём новый веб-вервер
	app := fiber.New()

	// создаём группу "/api"
	groupApi := app.Group("/api")

	// создаём подгруппу "api/v1"
	groupApiV1 := groupApi.Group("/v1")

	groupInternal := groupApi.Group("/internal")
	app.Get("/swagger/*", swagger.HandlerDefault)

	return &Server{
		App:           app,
		GroupApiV1:    groupApiV1,
		GroupInternal: groupInternal,
	}
}
