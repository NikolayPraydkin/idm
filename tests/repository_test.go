package tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	writeDotEnv("DB_DRIVER_NAME=postgres\nDB_DSN='host=127.0.0.1 port=5432 user=postgres password=postgres dbname=idm_tests sslmode=disable'")
	common.GetConfig(".env")
	// Перед всеми тестами: создаём БД idm_tests (если нужно).
	if _, err := CreateTestDB(); err != nil {
		panic(fmt.Errorf("failed to create test database: %v", err))
	}

	var err error
	err = CreateEmployeeTestTable()
	if err != nil {
		panic(fmt.Errorf("failed to connect to test database: %v", err))
	}
	err = CreateRoleTestTable()
	if err != nil {
		panic(fmt.Errorf("failed to connect to test database: %v", err))
	}
	// Запускаем все тесты в этом пакете:
	code := m.Run()
	removeDotEnv()
	os.Exit(code)
}

func TestRepository(t *testing.T) {
	a := assert.New(t)
	var db = database.ConnectDb()
	var clearDatabase = func() {
		db.MustExec("delete from employee")
	}
	defer func() {
		if r := recover(); r != nil {
			clearDatabase()
		}
	}()
	var employeeRepository = employee.NewEmployeeRepository(db)
	var fixture = NewFixture(employeeRepository)

	t.Run("find an employee by id", func(t *testing.T) {
		var newEmployeeId = fixture.Employee("Test Name")

		got, err := employeeRepository.FindById(newEmployeeId)

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got.Id)
		a.NotEmpty(got.CreatedAt)
		a.NotEmpty(got.UpdatedAt)
		a.Equal("Test Name", got.Name)
		clearDatabase()
	})
}
