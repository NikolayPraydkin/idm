package tests

import (
	"github.com/jmoiron/sqlx"
	"idm/inner/employee"
	"os"
)

var testDB *sqlx.DB

type Fixture struct {
	employees *employee.EmployeeRepository
}

func NewFixture(employees *employee.EmployeeRepository) *Fixture {
	return &Fixture{employees}
}

func (f *Fixture) Employee(name string) int64 {
	var entity = employee.EmployeeEntity{
		Name: name,
	}
	var newId, err = f.employees.Save(&entity)
	if err != nil {
		panic(err)
	}
	return newId
}

func CreateTestDB() (*sqlx.DB, error) {
	adminDSN := os.Getenv("DB_DSN")
	if adminDSN == "" {
		adminDSN = "host=localhost port=5432 user=postgres password=postgres dbname=idm_tests sslmode=disable"
	}
	adminDB, err := sqlx.Connect("postgres", adminDSN)
	if err != nil {
		return nil, err
	}
	//defer adminDB.Close()

	// Проверяем, существует ли уже база idm_tests:
	var exists bool
	err = adminDB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", "idm_tests")
	if err != nil {
		return nil, err
	}
	if !exists {
		if _, err := adminDB.Exec("CREATE DATABASE idm_tests"); err != nil {
			return nil, err
		}
	}
	testDB = adminDB
	return adminDB, nil
}

func CreateEmployeeTestTable() error {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=idm_tests sslmode=disable"
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return err
	}

	// Создаём схему таблицы employee, если она отсутствует:
	schema := `
CREATE TABLE IF NOT EXISTS employee (
  id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name       TEXT        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Закрываем подключение к idm_tests:
	err = db.Close()
	if err != nil {
		return err
	}
	return nil
}
func CreateRoleTestTable() error {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=idm_tests sslmode=disable"
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return err
	}

	// Создаём схему таблицы employee, если она отсутствует:
	schema := `
CREATE TABLE IF NOT EXISTS role
(	id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Закрываем подключение к idm_tests:
	err = db.Close()
	if err != nil {
		return err
	}
	return nil
}
