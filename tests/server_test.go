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
