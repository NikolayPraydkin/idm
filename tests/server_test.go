package tests

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"idm/inner/server"
	"io"
	"net/http/httptest"
	"testing"
)

func Test_EmployeePaginationIntegration(t *testing.T) {
	TruncateTable(testDB)
	db, _ := CreateTestDB()

	// Записываем 5 записей
	for i := 1; i <= 5; i++ {
		_, err := db.Exec("INSERT INTO employee (name) VALUES ($1)", "user"+string(rune('A'+i)))
		assert.NoError(t, err)
	}

	tests := []struct {
		name       string
		query      string
		wantCount  int
		statusCode int
	}{
		{"Page 1", "/api/v1/employees/page?pageNumber=0&pageSize=3", 3, 200},
		{"Page 2", "/api/v1/employees/page?pageNumber=1&pageSize=3", 2, 200},
		{"Page 3", "/api/v1/employees/page?pageNumber=2&pageSize=3", 0, 200},
		{"Invalid PageSize", "/api/v1/employees/page?pageNumber=1&pageSize=0", 0, 400},
		{"Missing PageNumber", "/api/v1/employees/page?pageSize=3", 3, 200}, // предполагаем дефолт PageNumber=0
		{"Missing PageSize", "/api/v1/employees/page?pageNumber=0", 0, 400}, // считаем, что pageSize обязателен
	}

	srv, _ := server.Build()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.query, nil)
			resp, err := srv.App.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, tt.statusCode, resp.StatusCode)

			if tt.statusCode == 200 {
				body, _ := io.ReadAll(resp.Body)
				var response struct {
					Success bool   `json:"success"`
					Message string `json:"error"`
					Data    struct {
						Result []struct {
							ID        int    `json:"id"`
							Name      string `json:"name"`
							CreatedAt string `json:"created_at"`
							UpdatedAt string `json:"updated_at"`
						} `json:"result"`
						PageSize   int   `json:"page_size"`
						PageNumber int   `json:"page_number"`
						Total      int64 `json:"total"`
					} `json:"data"`
				}
				_ = json.Unmarshal(body, &response)
				assert.Equal(t, tt.wantCount, len(response.Data.Result))
			}
		})
	}

	testsWithError := []struct {
		name         string
		query        string
		statusCode   int
		errorMessage string
	}{
		{"BadRequst", "/api/v1/employees/page?pageNumber=2&pageSize=bad", 400, "bad query params"},
		{"pageSizeMin", "/api/v1/employees/page?pageNumber=0&pageSize=0", 400, "Key: 'PageRequest.PageSize' Error:Field validation for 'PageSize' failed on the 'min' tag"},
		{"pageSizeMax", "/api/v1/employees/page?pageNumber=0&pageSize=101", 400, "Key: 'PageRequest.PageSize' Error:Field validation for 'PageSize' failed on the 'max' tag"},
	}

	for _, tt := range testsWithError {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.query, nil)
			resp, _ := srv.App.Test(req, -1)
			assert.Equal(t, tt.statusCode, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			var response struct {
				Success bool   `json:"success"`
				Message string `json:"error"`
			}
			_ = json.Unmarshal(body, &response)
			assert.Equal(t, tt.errorMessage, response.Message)
		})
	}

}

func Test_EmployeeSearchByNameIntegration(t *testing.T) {
	TruncateTable(testDB)
	db, _ := CreateTestDB()

	// Фиксируем набор данных для поиска
	names := []string{"Alice", "ALINA", "Bob", "alex", "Maria"}
	for _, n := range names {
		_, err := db.Exec("INSERT INTO employee (name) VALUES ($1)", n)
		assert.NoError(t, err)
	}

	srv, _ := server.Build()

	type respDTO struct {
		Success bool   `json:"success"`
		Message string `json:"error"`
		Data    struct {
			Result []struct {
				ID        int    `json:"id"`
				Name      string `json:"name"`
				CreatedAt string `json:"created_at"`
				UpdatedAt string `json:"updated_at"`
			} `json:"result"`
			PageSize   int   `json:"page_size"`
			PageNumber int   `json:"page_number"`
			Total      int64 `json:"total"`
		} `json:"data"`
	}

	// Предполагаем, что имя параметра фильтра — textFilter (как в PageRequest.TextFilter)
	// 1) нет фильтра / пустая строка / только пробелы и управляющие символы => фильтр игнорится
	t.Run("NoTextFilterParam", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=0&pageSize=50", nil)
		resp, err := srv.App.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		var r respDTO
		_ = json.Unmarshal(body, &r)
		assert.Equal(t, len(names), len(r.Data.Result))
	})

	t.Run("EmptyString", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=0&pageSize=50&textFilter=", nil)
		resp, err := srv.App.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		var r respDTO
		_ = json.Unmarshal(body, &r)
		assert.Equal(t, len(names), len(r.Data.Result))
	})

	t.Run("WhitespaceOnly", func(t *testing.T) {
		// три пробела
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=0&pageSize=50&textFilter=%20%20%20", nil)
		resp, err := srv.App.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		var r respDTO
		_ = json.Unmarshal(body, &r)
		assert.Equal(t, len(names), len(r.Data.Result))
	})

	t.Run("WhitespaceTabsNewlines", func(t *testing.T) {
		// таб и перевод строки: %09 = \t, %0A = \n
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=0&pageSize=50&textFilter=%09%0A", nil)
		resp, err := srv.App.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		var r respDTO
		_ = json.Unmarshal(body, &r)
		assert.Equal(t, len(names), len(r.Data.Result))
	})

	// 2) фильтр короче 3 символов => считаем, что он игнорируется
	t.Run("FilterLessThan3", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=0&pageSize=50&textFilter=al", nil)
		resp, err := srv.App.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		var r respDTO
		_ = json.Unmarshal(body, &r)
		assert.Equal(t, len(names), len(r.Data.Result))
	})

	// 3) фильтр от 3 символов => применяется регистронезависимый поиск подстроки
	t.Run("FilterAtLeast3", func(t *testing.T) {
		// "ali" должен найти "Alice" и "ALINA" (2 совпадения)
		req := httptest.NewRequest("GET", "/api/v1/employees/page?pageNumber=0&pageSize=50&textFilter=ali", nil)
		resp, err := srv.App.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		var r respDTO
		_ = json.Unmarshal(body, &r)
		assert.Equal(t, 2, len(r.Data.Result))
		// Дополнительно можно убедиться, что вернулись именно нужные имена
		got := make(map[string]bool)
		for _, e := range r.Data.Result {
			got[e.Name] = true
		}
		assert.True(t, got["Alice"])
		assert.True(t, got["ALINA"])
	})
}
