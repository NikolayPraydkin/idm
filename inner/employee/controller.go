package employee

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"
)

type Controller struct {
	server          *web.Server
	employeeService Svc
	logger          *common.Logger
}

// Svc описывает набор методов бизнес-логики по работе с сотрудниками
type Svc interface {
	FindById(id int64) (Response, error)
	Add(req CreateRequest) error
	Save(req CreateRequest) (int64, error)
	FindAll() ([]Response, error)
	FindByIds(ids []int64) ([]Response, error)
	DeleteById(id int64) error
	DeleteByIds(ids []int64) error
	SaveWithTransaction(e CreateRequest) (int64, error)
}

func NewController(server *web.Server, employeeService Svc, logger *common.Logger) *Controller {
	return &Controller{
		server:          server,
		employeeService: employeeService,
		logger:          logger,
	}
}

func (c *Controller) RegisterRoutes() {

	c.server.GroupApiV1.Post("/employees", c.CreateEmployee)
	c.server.GroupApiV1.Post("/employees/add", c.AddEmployee)
	c.server.GroupApiV1.Post("/employees/save", c.SaveEmployee)
	c.server.GroupApiV1.Get("/employees/:id", c.GetEmployee)
	c.server.GroupApiV1.Get("/employees", c.GetAllEmployees)
	c.server.GroupApiV1.Post("/employees/batch", c.GetEmployeesByIds)
	c.server.GroupApiV1.Delete("/employees/:id", c.DeleteEmployeeById)
	c.server.GroupApiV1.Delete("/employees", c.DeleteEmployeesByIds)
}

func (c *Controller) CreateEmployee(ctx *fiber.Ctx) error {

	// анмаршалим JSON body запроса в структуру CreateRequest
	var request CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("create employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("create employee: received request", zap.Any("request", request))

	// вызываем метод CreateEmployee сервиса employee.Service
	var newEmployeeId, err = c.employeeService.SaveWithTransaction(request)
	if err != nil {
		switch {
		// если сервис возвращает ошибку RequestValidationError или AlreadyExistsError,
		// то мы возвращаем ответ с кодом 400 (BadRequest)
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			c.logger.Error("create employee", zap.Error(err))
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		// если сервис возвращает другую ошибку, то мы возвращаем ответ с кодом 500 (InternalServerError)
		default:
			c.logger.Error("create employee", zap.Error(err))
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}

	// функция OkResponse() формирует и направляет ответ в случае успеха
	if err = common.OkResponse(ctx, newEmployeeId); err != nil {
		c.logger.Error("create employee", zap.Error(err))
		// функция ErrorResponse() формирует и направляет ответ в случае ошибки
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning created employee id")
	}

	return nil
}

// AddEmployee handles POST /api/v1/employees/add
func (c *Controller) AddEmployee(ctx *fiber.Ctx) error {
	var req CreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		c.logger.Error("Add employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("Add employee: received request", zap.Any("request", req))

	if err := c.employeeService.Add(req); err != nil {
		switch {
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			c.logger.Error("Add employee", zap.Error(err))
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			c.logger.Error("Add employee", zap.Error(err))
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return common.OkResponse(ctx, fiber.Map{"message": "added"})
}

// SaveEmployee handles POST /api/v1/employees/save
func (c *Controller) SaveEmployee(ctx *fiber.Ctx) error {
	var req CreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		c.logger.Error("Save employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("Save employee: received request", zap.Any("request", req))

	id, err := c.employeeService.Save(req)
	if err != nil {
		switch {
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			c.logger.Error("Save employee", zap.Error(err))
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			c.logger.Error("Save employee", zap.Error(err))
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return common.OkResponse(ctx, fiber.Map{"id": id})
}

// GetEmployee handles GET /api/v1/employees/:id
func (c *Controller) GetEmployee(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	c.logger.Debug("Get employee", zap.String("id", idParam))

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.logger.Error("Get employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid id")
	}
	resp, err := c.employeeService.FindById(id)
	if err != nil {
		c.logger.Error("Get employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, resp)
}

// GetAllEmployees handles GET /api/v1/employees
func (c *Controller) GetAllEmployees(ctx *fiber.Ctx) error {
	resps, err := c.employeeService.FindAll()
	if err != nil {
		c.logger.Error("Get all employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, resps)
}

// GetEmployeesByIds handles POST /api/v1/employees/batch
func (c *Controller) GetEmployeesByIds(ctx *fiber.Ctx) error {
	var ids []int64
	if err := ctx.BodyParser(&ids); err != nil {
		c.logger.Error("Get employees by ids", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	resps, err := c.employeeService.FindByIds(ids)
	if err != nil {
		c.logger.Error("Get all employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, resps)
}

// DeleteEmployeeById handles DELETE /api/v1/employees/:id
func (c *Controller) DeleteEmployeeById(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	c.logger.Debug("Delete employee", zap.String("id", idParam))
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.logger.Error("Delete employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "invalid id")
	}
	if err := c.employeeService.DeleteById(id); err != nil {
		c.logger.Error("Delete employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, fiber.Map{"message": "deleted"})
}

// DeleteEmployeesByIds handles DELETE /api/v1/employees
func (c *Controller) DeleteEmployeesByIds(ctx *fiber.Ctx) error {
	var ids []int64
	if err := ctx.BodyParser(&ids); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("Delete employees by ids", zap.Int64s("ids", ids))
	if err := c.employeeService.DeleteByIds(ids); err != nil {
		c.logger.Error("Delete employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, fiber.Map{"message": "deleted"})
}
