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
	GetEmployeesPage(req PageRequest) (PageResponse, error)
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
	c.server.GroupApiV1.Get("/employees", c.GetAllEmployees)
	c.server.GroupApiV1.Delete("/employees", c.DeleteEmployeesByIds)
	c.server.GroupApiV1.Get("/employees/page", c.GetEmployeesPage)
	c.server.GroupApiV1.Post("/employees/batch", c.GetEmployeesByIds)
	c.server.GroupApiV1.Get("/employees/:id", c.GetEmployee)
	c.server.GroupApiV1.Delete("/employees/:id", c.DeleteEmployeeById)
}

// Функция-хендлер, которая будет вызываться при POST запросе по маршруту "/api/v1/employees"
// @Description Create a new employee.
// @Summary create a new employee
// @Tags employee
// @Accept json
// @Produce json
// @Param request body employee.CreateRequest true "create employee request"
// @Success 200 {object} common.Response[int64]
// @Router /employees [post]
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

// AddEmployee godoc
// @Summary      Add employee
// @Description  Add a new employee (no id returned)
// @Tags         employee
// @Accept       json
// @Produce      json
// @Param        request  body      employee.CreateRequest  true  "create employee request"
// @Success      200      {object}  map[string]string
// @Router       /employees/add [post]
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

// SaveEmployee godoc
// @Summary      Save employee
// @Description  Create or update employee and return its id
// @Tags         employee
// @Accept       json
// @Produce      json
// @Param        request  body      employee.CreateRequest  true  "save employee request"
// @Success      200      {object}  map[string]int64
// @Router       /employees/save [post]
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

// GetEmployee godoc
// @Summary      Get employee by id
// @Description  Returns employee by id
// @Tags         employee
// @Produce      json
// @Param        id   path      int  true  "employee id"
// @Success      200  {object}  employee.Response
// @Router       /employees/{id} [get]
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

// GetAllEmployees godoc
// @Summary      List employees
// @Description  Returns all employees
// @Tags         employee
// @Produce      json
// @Success      200  {array}  employee.Response
// @Router       /employees [get]
// GetAllEmployees handles GET /api/v1/employees
func (c *Controller) GetAllEmployees(ctx *fiber.Ctx) error {
	resps, err := c.employeeService.FindAll()
	if err != nil {
		c.logger.Error("Get all employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, resps)
}

// GetEmployeesPage godoc
// @Summary      Get employees page
// @Description  Returns paginated list of employees
// @Tags         employee
// @Produce      json
// @Param        request  query     employee.PageRequest  true  "page request"
// @Success      200      {object}  employee.PageResponse
// @Router       /employees/page [get]
func (c *Controller) GetEmployeesPage(ctx *fiber.Ctx) error {
	var req PageRequest
	if err := ctx.QueryParser(&req); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "bad query params")
	}
	pageResp, err := c.employeeService.GetEmployeesPage(req)
	if err != nil {
		if errors.As(err, &common.RequestValidationError{}) {
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		}
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, pageResp)
}

// GetEmployeesByIds godoc
// @Summary      Get employees by ids
// @Description  Returns employees by the given ids
// @Tags         employee
// @Accept       json
// @Produce      json
// @Param        ids  body      []int64  true  "employee ids"
// @Success      200  {array}   employee.Response
// @Router       /employees/batch [post]
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

// DeleteEmployeeById godoc
// @Summary      Delete employee by id
// @Description  Deletes employee with the specified id
// @Tags         employee
// @Produce      json
// @Param        id   path      int  true  "employee id"
// @Success      200  {object}  map[string]string
// @Router       /employees/{id} [delete]
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

// DeleteEmployeesByIds godoc
// @Summary      Delete employees by ids
// @Description  Deletes employees with the specified ids
// @Tags         employee
// @Accept       json
// @Produce      json
// @Param        ids  body      []int64  true  "employee ids"
// @Success      200  {object}  map[string]string
// @Router       /employees [delete]
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
