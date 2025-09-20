package web

import (
	_ "idm/docs"

	"github.com/gofiber/fiber/v2"
)

// структуа веб-сервера
type Server struct {
	App *fiber.App
	// группа публичного API
	GroupApi fiber.Router
	// группа публичного API первой версии
	GroupApiV1 fiber.Router
	// группа непубличного API
	GroupInternal fiber.Router
}

type AuthMiddlewareInterface interface {
	ProtectWithJwt() func(*fiber.Ctx) error
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

	return &Server{
		App:           app,
		GroupApi:      groupApi,
		GroupApiV1:    groupApiV1,
		GroupInternal: groupInternal,
	}
}
