package info

import (
	"github.com/gofiber/fiber/v2"
	"idm/inner/common"
	"idm/inner/web"
)

type Controller struct {
	server *web.Server
	cfg    common.Config
}

func NewController(server *web.Server, cfg common.Config) *Controller {
	return &Controller{
		server: server,
		cfg:    cfg,
	}
}

type InfoResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (c *Controller) RegisterRoutes() {
	c.server.GroupInternal.Get("/info", c.GetInfo)
	c.server.GroupInternal.Get("/health", c.GetHealth)
}

func (c *Controller) GetInfo(ctx *fiber.Ctx) error {
	var err = ctx.Status(fiber.StatusOK).JSON(&InfoResponse{
		Name:    c.cfg.AppName,
		Version: c.cfg.AppVersion,
	})
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning info")
	}
	return nil
}

func (c *Controller) GetHealth(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).SendString("OK")
}
