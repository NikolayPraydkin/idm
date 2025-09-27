package employee

import (
	"context"
	"encoding/json"
	"errors"
	"idm/inner/common"
	"idm/inner/web"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Объявляем структуру мока сервиса employee.Service
type MockService struct {
	mock.Mock
}

func (svc *MockService) Add(ctx context.Context, req CreateRequest) error {
	args := svc.Called(req)
	return args.Error(0)
}

func (svc *MockService) Save(ctx context.Context, req CreateRequest) (int64, error) {
	args := svc.Called(req)
	return args.Get(0).(int64), args.Error(1)
}

func (svc *MockService) FindAll(ctx context.Context) ([]Response, error) {
	args := svc.Called()
	return args.Get(0).([]Response), args.Error(1)
}

func (svc *MockService) FindByIds(ctx context.Context, ids []int64) ([]Response, error) {
	args := svc.Called(ids)
	return args.Get(0).([]Response), args.Error(1)
}

func (svc *MockService) DeleteById(ctx context.Context, id int64) error {
	args := svc.Called(id)
	return args.Error(0)
}

func (svc *MockService) DeleteByIds(ctx context.Context, ids []int64) error {
	args := svc.Called(ids)
	return args.Error(0)
}

func (svc *MockService) SaveWithTransaction(ctx context.Context, e CreateRequest) (int64, error) {
	args := svc.Called(e.ToEntity())
	return args.Get(0).(int64), args.Error(1)
}

// Реализуем функции мок-сервиса
func (svc *MockService) FindById(ctx context.Context, id int64) (Response, error) {
	args := svc.Called(id)
	return args.Get(0).(Response), args.Error(1)
}

func (svc *MockService) CreateEmployee(ctx context.Context, request CreateRequest) (int64, error) {
	args := svc.Called(request)
	return args.Get(0).(int64), args.Error(1)
}

func (svc *MockService) GetEmployeesPage(ctx context.Context, req PageRequest) (PageResponse, error) {
	args := svc.Called(req)
	return args.Get(0).(PageResponse), args.Error(1)
}

func TestCreateEmployee(t *testing.T) {
	a := assert.New(t)

	tests := []struct {
		name       string
		body       string
		mockSetup  func(*MockService)
		wantStatus int
		wantID     int64
	}{
		{
			name: "should return created employee id",
			body: `{"name": "john doe"}`,
			mockSetup: func(svc *MockService) {
				svc.On("SaveWithTransaction", mock.AnythingOfType("*employee.Entity")).Return(int64(123), nil)
			},
			wantStatus: http.StatusOK,
			wantID:     123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаём тестовый токен аутентификации с ролью web.IdmAdmin
			var claims = &web.IdmClaims{
				RealmAccess: web.RealmAccessClaims{Roles: []string{web.IdmAdmin}},
			}
			// создаём stub middleware для аутентификации
			var auth = func(c *fiber.Ctx) error {
				c.Locals(web.JwtKey, &jwt.Token{Claims: claims})
				return c.Next()
			}

			server := web.NewServer()
			server.GroupApiV1.Use(auth)

			svc := new(MockService)
			logger := common.NewLogger(common.GetConfig(".env"))
			controller := NewController(server, svc, logger)
			controller.RegisterRoutes()

			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := server.App.Test(req)
			a.Nil(err)
			a.NotEmpty(resp)
			a.Equal(tt.wantStatus, resp.StatusCode)

			bytesData, err := io.ReadAll(resp.Body)
			a.Nil(err)

			var responseBody common.Response[int64]
			err = json.Unmarshal(bytesData, &responseBody)
			a.Nil(err)
			a.Equal(tt.wantID, responseBody.Data)
			a.True(responseBody.Success)
			a.Empty(responseBody.Message)
		})
	}
}

func TestAddEmployee(t *testing.T) {
	a := assert.New(t)

	tests := []struct {
		name       string
		body       string
		mockSetup  func(*MockService)
		wantStatus int
	}{
		{
			name: "should add employee successfully",
			body: `{"name":"alice"}`,
			mockSetup: func(svc *MockService) {
				svc.On("Add", mock.AnythingOfType("employee.CreateRequest")).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return bad request on validation error",
			body: `{}`,
			mockSetup: func(svc *MockService) {
				svc.On("Add", mock.AnythingOfType("employee.CreateRequest")).Return(common.RequestValidationError{Message: "Add emploee"})
			},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "should return internal error on service failure",
			body: `{"name":"alice"}`,
			mockSetup: func(svc *MockService) {
				svc.On("Add", mock.AnythingOfType("employee.CreateRequest")).Return(errors.New("fail"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаём тестовый токен аутентификации с ролью web.IdmAdmin
			var claims = &web.IdmClaims{
				RealmAccess: web.RealmAccessClaims{Roles: []string{web.IdmAdmin}},
			}
			// создаём stub middleware для аутентификации
			var auth = func(c *fiber.Ctx) error {
				c.Locals(web.JwtKey, &jwt.Token{Claims: claims})
				return c.Next()
			}

			server := web.NewServer()
			server.GroupApiV1.Use(auth)

			svc := new(MockService)
			controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
			controller.RegisterRoutes()

			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/add", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := server.App.Test(req)
			a.NoError(err)
			a.Equal(tt.wantStatus, resp.StatusCode)

			// Для успешного кейса дополнительно проверим тело
			if tt.wantStatus == http.StatusOK {
				data, _ := io.ReadAll(resp.Body)
				var rb common.Response[map[string]string]
				a.NoError(json.Unmarshal(data, &rb))
				a.True(rb.Success)
				a.Equal("added", rb.Data["message"])
			}
		})
	}
}

func TestSaveEmployee(t *testing.T) {
	a := assert.New(t)

	tests := []struct {
		name       string
		body       string
		mockSetup  func(*MockService)
		wantStatus int
		wantID     int64
		checkBody  bool
	}{
		{
			name: "should save employee and return id",
			body: `{"name":"bob"}`,
			mockSetup: func(svc *MockService) {
				svc.On("Save", mock.AnythingOfType("employee.CreateRequest")).Return(int64(42), nil)
			},
			wantStatus: http.StatusOK,
			wantID:     42,
			checkBody:  true,
		},
		{
			name: "should return bad request on validation error",
			body: `{}`,
			mockSetup: func(svc *MockService) {
				svc.On("Save", mock.AnythingOfType("employee.CreateRequest")).Return(int64(0), common.RequestValidationError{Message: "Save employee"})
			},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "should return internal error on service failure",
			body: `{"name":"bob"}`,
			mockSetup: func(svc *MockService) {
				svc.On("Save", mock.AnythingOfType("employee.CreateRequest")).Return(int64(0), errors.New("fail"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаём тестовый токен аутентификации с ролью web.IdmAdmin
			var claims = &web.IdmClaims{
				RealmAccess: web.RealmAccessClaims{Roles: []string{web.IdmAdmin}},
			}
			// создаём stub middleware для аутентификации
			var auth = func(c *fiber.Ctx) error {
				c.Locals(web.JwtKey, &jwt.Token{Claims: claims})
				return c.Next()
			}

			server := web.NewServer()
			server.GroupApiV1.Use(auth)

			svc := new(MockService)
			controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
			controller.RegisterRoutes()

			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/save", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := server.App.Test(req)
			a.NoError(err)
			a.Equal(tt.wantStatus, resp.StatusCode)

			if tt.checkBody {
				data, _ := io.ReadAll(resp.Body)
				var rb common.Response[map[string]int64]
				a.NoError(json.Unmarshal(data, &rb))
				a.True(rb.Success)
				a.Equal(tt.wantID, rb.Data["id"])
			}
		})
	}
}

func TestGetEmployee(t *testing.T) {
	a := assert.New(t)

	tests := []struct {
		name       string
		url        string
		mockSetup  func(*MockService)
		wantStatus int
		wantID     int64
		checkBody  bool
	}{
		{
			name: "should return employee",
			url:  "/api/v1/employees/7",
			mockSetup: func(svc *MockService) {
				svc.On("FindById", int64(7)).Return(Response{Id: 7, Name: "E"}, nil)
			},
			wantStatus: http.StatusOK,
			wantID:     7,
			checkBody:  true,
		},
		{
			name:       "should return bad request on invalid id",
			url:        "/api/v1/employees/abc",
			mockSetup:  nil,
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "should return internal error on service failure",
			url:  "/api/v1/employees/8",
			mockSetup: func(svc *MockService) {
				svc.On("FindById", int64(8)).Return(Response{}, errors.New("fail"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаём тестовый токен аутентификации с ролью web.IdmAdmin
			var claims = &web.IdmClaims{
				RealmAccess: web.RealmAccessClaims{Roles: []string{web.IdmAdmin}},
			}
			// создаём stub middleware для аутентификации
			var auth = func(c *fiber.Ctx) error {
				c.Locals(web.JwtKey, &jwt.Token{Claims: claims})
				return c.Next()
			}

			server := web.NewServer()
			server.GroupApiV1.Use(auth)

			svc := new(MockService)
			controller := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
			controller.RegisterRoutes()

			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			req := httptest.NewRequest(fiber.MethodGet, tt.url, nil)
			resp, err := server.App.Test(req)
			a.NoError(err)
			a.Equal(tt.wantStatus, resp.StatusCode)

			if tt.checkBody {
				data, _ := io.ReadAll(resp.Body)
				var rb common.Response[Response]
				a.NoError(json.Unmarshal(data, &rb))
				a.True(rb.Success)
				a.Equal(tt.wantID, rb.Data.Id)
			}
		})
	}
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
			// создаём тестовый токен аутентификации с ролью web.IdmAdmin
			var claims = &web.IdmClaims{
				RealmAccess: web.RealmAccessClaims{Roles: []string{web.IdmAdmin}},
			}
			// создаём stub middleware для аутентификации
			var auth = func(c *fiber.Ctx) error {
				c.Locals(web.JwtKey, &jwt.Token{Claims: claims})
				return c.Next()
			}

			server := web.NewServer()
			server.GroupApiV1.Use(auth)

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

func TestAuthorization_AdminEndpoints(t *testing.T) {
	a := assert.New(t)
	secret := []byte("test-secret")

	tests := []struct {
		name           string
		authHeader     string
		setupMocks     func(*MockService)
		wantStatusCode int
	}{
		{
			name:           "no token -> 401",
			authHeader:     "",
			setupMocks:     nil,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "malformed token -> 401",
			authHeader:     "Bearer i.am.not.a.jwt",
			setupMocks:     nil,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "expired token -> 401",
			authHeader: func() string {
				s, _ := makeHS256Token(secret, []string{web.IdmAdmin}, true, true)
				return "Bearer " + s
			}(),
			setupMocks:     nil,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "bad signature -> 401",
			authHeader: func() string {
				// подпишем другим ключом
				s, _ := makeHS256Token([]byte("wrong-secret"), []string{web.IdmAdmin}, true, false)
				return "Bearer " + s
			}(),
			setupMocks:     nil,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "valid token but missing role -> 403",
			authHeader: func() string {
				s, _ := makeHS256Token(secret, []string{web.IdmUser}, true, false)
				return "Bearer " + s
			}(),
			setupMocks:     nil,
			wantStatusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := web.NewServer()
			// Навешиваем тестовый JWT-мидлвар
			server.GroupApiV1.Use(testJWTMiddleware(secret))

			svc := new(MockService)
			logger := common.NewLogger(common.GetConfig(".env"))
			ctrl := NewController(server, svc, logger)
			ctrl.RegisterRoutes()

			if tt.setupMocks != nil {
				tt.setupMocks(svc)
			}

			req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees", strings.NewReader(`{"name":"x"}`))
			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := server.App.Test(req, -1)
			a.NoError(err)
			a.Equal(tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func TestAuthorization_ReadEndpoints(t *testing.T) {
	a := assert.New(t)
	secret := []byte("test-secret")

	tests := []struct {
		name           string
		authHeader     string
		setupMocks     func(*MockService)
		wantStatusCode int
	}{
		{
			name:           "no token -> 401",
			authHeader:     "",
			setupMocks:     nil,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "malformed token -> 401",
			authHeader:     "Bearer nope",
			setupMocks:     nil,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "expired token -> 401",
			authHeader: func() string {
				s, _ := makeHS256Token(secret, []string{web.IdmUser}, true, true)
				return "Bearer " + s
			}(),
			setupMocks:     nil,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "bad signature -> 401",
			authHeader: func() string {
				s, _ := makeHS256Token([]byte("wrong-secret"), []string{web.IdmUser}, true, false)
				return "Bearer " + s
			}(),
			setupMocks:     nil,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "valid token but missing roles (empty) -> 403",
			authHeader: func() string {
				s, _ := makeHS256Token(secret, []string{}, true, false)
				return "Bearer " + s
			}(),
			setupMocks:     nil,
			wantStatusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := web.NewServer()
			server.GroupApiV1.Use(testJWTMiddleware(secret))

			svc := new(MockService)
			ctrl := NewController(server, svc, common.NewLogger(common.GetConfig(".env")))
			ctrl.RegisterRoutes()

			// мок для успешного чтения, он не должен вызываться при 401/403
			svc.On("FindById", int64(7)).Return(Response{Id: 7, Name: "E"}, nil).Maybe()

			req := httptest.NewRequest(fiber.MethodGet, "/api/v1/employees/7", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := server.App.Test(req, -1)
			a.NoError(err)
			a.Equal(tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func testJWTMiddleware(secret []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			// не ставим locals — дальше requireRoles вернёт 401
			return c.Next()
		}
		raw := strings.TrimSpace(auth[len("Bearer "):])

		claims := &web.IdmClaims{}
		// Валидация подписи и exp через стандартный парсер jwt/v5
		tkn, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (interface{}, error) {
			// принимаем только HS256 для тестов
			if token.Method.Alg() != "HS256" {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		})
		if err != nil || !tkn.Valid {
			// токен невалидный — locals не ставим
			return c.Next()
		}
		// токен валиден — кладём в Locals для requireRoles(...)
		c.Locals(web.JwtKey, tkn)
		return c.Next()
	}
}

func makeHS256Token(secret []byte, roles []string, withExp bool, expired bool) (string, error) {
	claims := web.IdmClaims{
		RealmAccess: web.RealmAccessClaims{Roles: roles},
	}

	// Пытаемся через рефлексию безопасно установить exp, если поле есть
	if withExp {
		// Устанавливаем exp на основе наличия RegisteredClaims внутри
		switch v := any(&claims).(type) {
		case *web.IdmClaims:
			// Пытаемся заполнить через встроенные поля, если есть
			// (если RegisteredClaims отсутсвуют, просто игнорируем — токен будет без exp)
			_ = v
		}
	}

	now := time.Now()
	var expiresAt *jwt.NumericDate
	if withExp {
		if expired {
			expiresAt = jwt.NewNumericDate(now.Add(-1 * time.Hour))
		} else {
			expiresAt = jwt.NewNumericDate(now.Add(1 * time.Hour))
		}
	}

	// Сконструируем map claims, чтобы гарантированно положить exp, не завися от структуры IdmClaims
	m := jwt.MapClaims{
		"realm_access": map[string]any{"roles": roles},
		"iat":          now.Unix(),
	}
	if withExp && expiresAt != nil {
		m["exp"] = expiresAt.Unix()
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, m)
	return t.SignedString(secret)
}
