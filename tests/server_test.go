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

	type testCase struct {
		name          string
		query         string
		wantStatus    int
		wantCount     int
		wantNamesOnly []string
	}

	tests := []testCase{
		{
			name:       "NoTextFilterParam",
			query:      "/api/v1/employees/page?pageNumber=0&pageSize=50",
			wantStatus: 200,
			wantCount:  len(names),
		},
		{
			name:       "EmptyString",
			query:      "/api/v1/employees/page?pageNumber=0&pageSize=50&textFilter=",
			wantStatus: 200,
			wantCount:  len(names),
		},
		{
			name:       "WhitespaceOnly",
			query:      "/api/v1/employees/page?pageNumber=0&pageSize=50&textFilter=%20%20%20",
			wantStatus: 200,
			wantCount:  len(names),
		},
		{
			name:       "WhitespaceTabsNewlines",
			query:      "/api/v1/employees/page?pageNumber=0&pageSize=50&textFilter=%09%0A",
			wantStatus: 200,
			wantCount:  len(names),
		},
		{
			name:       "FilterLessThan3",
			query:      "/api/v1/employees/page?pageNumber=0&pageSize=50&textFilter=al",
			wantStatus: 200,
			wantCount:  len(names),
		},
		{
			name:          "FilterAtLeast3",
			query:         "/api/v1/employees/page?pageNumber=0&pageSize=50&textFilter=ali",
			wantStatus:    200,
			wantCount:     2,
			wantNamesOnly: []string{"Alice", "ALINA"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.query, nil)
			resp, err := srv.App.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus != 200 {
				return
			}

			body, _ := io.ReadAll(resp.Body)
			var r respDTO
			_ = json.Unmarshal(body, &r)
			assert.Equal(t, tc.wantCount, len(r.Data.Result))

			if len(tc.wantNamesOnly) > 0 {
				got := make(map[string]bool)
				for _, e := range r.Data.Result {
					got[e.Name] = true
				}
				for _, expected := range tc.wantNamesOnly {
					assert.Truef(t, got[expected], "expected name %q not found in result", expected)
				}
			}
		})
	}
}
