package employee

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"idm/inner/common"
	"idm/inner/web"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Объявляем структуру мока сервиса employee.Service
type MockService struct {
	mock.Mock
}

func (svc *MockService) Add(req CreateRequest) error {
	args := svc.Called(req)
	return args.Error(0)
}

func (svc *MockService) Save(req CreateRequest) (int64, error) {
	args := svc.Called(req)
	return args.Get(0).(int64), args.Error(1)
}

func (svc *MockService) FindAll() ([]Response, error) {
	args := svc.Called()
	return args.Get(0).([]Response), args.Error(1)
}

func (svc *MockService) FindByIds(ids []int64) ([]Response, error) {
	args := svc.Called(ids)
	return args.Get(0).([]Response), args.Error(1)
}

func (svc *MockService) DeleteById(id int64) error {
	args := svc.Called(id)
	return args.Error(0)
}

func (svc *MockService) DeleteByIds(ids []int64) error {
	args := svc.Called(ids)
	return args.Error(0)
}

func (svc *MockService) SaveWithTransaction(e CreateRequest) (int64, error) {
	args := svc.Called(e.ToEntity())
	return args.Get(0).(int64), args.Error(1)
}

// Реализуем функции мок-сервиса
func (svc *MockService) FindById(id int64) (Response, error) {
	args := svc.Called(id)
	return args.Get(0).(Response), args.Error(1)
}

func (svc *MockService) CreateEmployee(request CreateRequest) (int64, error) {
	args := svc.Called(request)
	return args.Get(0).(int64), args.Error(1)
}

func (svc *MockService) GetEmployeesPage(req PageRequest) (PageResponse, error) {
	args := svc.Called(req)
	return args.Get(0).(PageResponse), args.Error(1)
}

func TestCreateEmployee(t *testing.T) {
	var a = assert.New(t)

	// тестируем положительный сценарий: работника создали и получили его id
	t.Run("should return created employee id", func(t *testing.T) {
		// Готовим тестовое окружение
		server := web.NewServer()
		var svc = new(MockService)
		var logger = common.NewLogger(common.GetConfig(".env"))
		var controller = NewController(server, svc, logger)
		controller.RegisterRoutes()
		// Готовим тестовое окружение
		var body = strings.NewReader("{\"name\": \"john doe\"}")
		var req = httptest.NewRequest(fiber.MethodPost, "/api/v1/employees", body)
		req.Header.Set("Content-Type", "application/json")

		// Настраиваем поведение мока в тесте
		svc.On("SaveWithTransaction", mock.AnythingOfType("*employee.Entity")).Return(int64(123), nil)

		// Отправляем тестовый запрос на веб сервер
		resp, err := server.App.Test(req)

		// Выполняем проверки полученных данных
		a.Nil(err)
		a.NotEmpty(resp)
		a.Equal(http.StatusOK, resp.StatusCode)
		bytesData, err := io.ReadAll(resp.Body)
		a.Nil(err)
		var responseBody common.Response[int64]
		err = json.Unmarshal(bytesData, &responseBody)
		a.Nil(err)
		a.Equal(int64(123), responseBody.Data)
		a.True(responseBody.Success)
		a.Empty(responseBody.Message)
	})
}

// TestAddEmployee tests the AddEmployee handler
func TestAddEmployee(t *testing.T) {
	a := assert.New(t)
	t.Run("should add employee successfully", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
		controller.RegisterRoutes()

		body := strings.NewReader(`{"name":"alice"}`)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/add", body)
		req.Header.Set("Content-Type", "application/json")

		svc.On("Add", mock.AnythingOfType("employee.CreateRequest")).Return(nil)

		resp, err := server.App.Test(req)
		a.NoError(err)
		a.Equal(http.StatusOK, resp.StatusCode)

		data, _ := io.ReadAll(resp.Body)
		var rb common.Response[map[string]string]
		a.NoError(json.Unmarshal(data, &rb))
		a.True(rb.Success)
		a.Equal("added", rb.Data["message"])
	})

	t.Run("should return bad request on validation error", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
		controller.RegisterRoutes()

		svc.On("Add", mock.AnythingOfType("employee.CreateRequest")).Return(common.RequestValidationError{Message: "Add emploee"})

		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/add", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req)
		a.NoError(err)
		a.Equal(fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return internal error on service failure", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
		controller.RegisterRoutes()

		body := strings.NewReader(`{"name":"alice"}`)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/add", body)
		req.Header.Set("Content-Type", "application/json")

		svc.On("Add", mock.AnythingOfType("employee.CreateRequest")).Return(errors.New("fail"))

		resp, err := server.App.Test(req)
		a.NoError(err)
		a.Equal(fiber.StatusInternalServerError, resp.StatusCode)
	})
}

// TestSaveEmployee tests the SaveEmployee handler
func TestSaveEmployee(t *testing.T) {
	a := assert.New(t)
	t.Run("should save employee and return id", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
		controller.RegisterRoutes()

		body := strings.NewReader(`{"name":"bob"}`)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/save", body)
		req.Header.Set("Content-Type", "application/json")

		svc.On("Save", mock.AnythingOfType("employee.CreateRequest")).Return(int64(42), nil)

		resp, err := server.App.Test(req)
		a.NoError(err)
		a.Equal(http.StatusOK, resp.StatusCode)

		data, _ := io.ReadAll(resp.Body)
		var rb common.Response[map[string]int64]
		a.NoError(json.Unmarshal(data, &rb))
		a.True(rb.Success)
		a.Equal(int64(42), rb.Data["id"])
	})

	t.Run("should return bad request on validation error", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
		controller.RegisterRoutes()

		svc.On("Save", mock.AnythingOfType("employee.CreateRequest")).Return(int64(0), common.RequestValidationError{Message: "Save employee"})

		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/save", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req)
		a.NoError(err)
		a.Equal(fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return internal error on service failure", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
		controller.RegisterRoutes()

		body := strings.NewReader(`{"name":"bob"}`)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/save", body)
		req.Header.Set("Content-Type", "application/json")

		svc.On("Save", mock.AnythingOfType("employee.CreateRequest")).Return(int64(0), errors.New("fail"))

		resp, err := server.App.Test(req)
		a.NoError(err)
		a.Equal(fiber.StatusInternalServerError, resp.StatusCode)
	})
}

// TestGetEmployee tests the GetEmployee handler
func TestGetEmployee(t *testing.T) {
	a := assert.New(t)
	t.Run("should return employee", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
		controller.RegisterRoutes()

		svc.On("FindById", int64(7)).Return(Response{Id: 7, Name: "E"}, nil)

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/employees/7", nil)
		resp, err := server.App.Test(req)
		a.NoError(err)
		a.Equal(http.StatusOK, resp.StatusCode)

		data, _ := io.ReadAll(resp.Body)
		var rb common.Response[Response]
		a.NoError(json.Unmarshal(data, &rb))
		a.True(rb.Success)
		a.Equal(int64(7), rb.Data.Id)
	})

	t.Run("should return bad request on invalid id", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
		controller.RegisterRoutes()

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/employees/abc", nil)
		resp, err := server.App.Test(req)
		a.NoError(err)
		a.Equal(fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return internal error on service failure", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
		controller.RegisterRoutes()

		svc.On("FindById", int64(8)).Return(Response{}, errors.New("fail"))

		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/employees/8", nil)
		resp, err := server.App.Test(req)
		a.NoError(err)
		a.Equal(fiber.StatusInternalServerError, resp.StatusCode)
	})
}

func TestGetEmployeesPageValidation(t *testing.T) {
	a := assert.New(t)

	tests := []struct {
		name       string
		query      string
		statusCode int
	}{
		{"page size too small", "?pageSize=0&pageNumber=1", fiber.StatusBadRequest},
		{"page size too large", "?pageSize=101&pageNumber=1", fiber.StatusBadRequest},
		{"page number negative", "?pageSize=10&pageNumber=-1", fiber.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := web.NewServer()
			repo := new(MockRepo)
			svc := newTestService(repo)
			controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
			controller.RegisterRoutes()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page"+tt.query, nil)
			resp, err := server.App.Test(req, -1)
			a.NoError(err)
			a.Equal(tt.statusCode, resp.StatusCode)
		})
	}
}
