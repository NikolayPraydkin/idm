package tests

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"idm/inner/employee"
	"testing"
	"time"
)

func TruncateTable(db *sqlx.DB) {
	db.Exec("TRUNCATE employee RESTART IDENTITY CASCADE")
}

func TestEmployeeRepository_AddAndFindById(t *testing.T) {
	TruncateTable(testDB)

	repo := employee.NewEmployeeRepository(testDB)
	now := time.Now()
	e := &employee.EmployeeEntity{
		Name:      "Alice",
		CreatedAt: now,
		UpdatedAt: now,
	}

	id, err := repo.Save(e) // Save возвращает (id, error)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := repo.FindById(id)
	if err != nil {
		t.Fatalf("FindById(%d) error = %v", id, err)
	}
	if got.Name != "Alice" {
		t.Errorf("FindById(%d).Name = %q; want %q", id, got.Name, "Alice")
	}
	TruncateTable(testDB)
}

func TestEmployeeRepository_FindAll(t *testing.T) {
	TruncateTable(testDB)

	repo := employee.NewEmployeeRepository(testDB)
	now := time.Now()
	_, _ = repo.Save(&employee.EmployeeEntity{Name: "Bob", CreatedAt: now, UpdatedAt: now})
	_, _ = repo.Save(&employee.EmployeeEntity{Name: "Carol", CreatedAt: now, UpdatedAt: now})

	all, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(all) != 2 {
		t.Errorf("FindAll() len = %d; want 2", len(all))
	}
}

func TestEmployeeRepository_FindByIds(t *testing.T) {
	TruncateTable(testDB)

	repo := employee.NewEmployeeRepository(testDB)
	now := time.Now()
	id1, _ := repo.Save(&employee.EmployeeEntity{Name: "Dave", CreatedAt: now, UpdatedAt: now})
	id2, _ := repo.Save(&employee.EmployeeEntity{Name: "Eve", CreatedAt: now, UpdatedAt: now})

	subset, err := repo.FindByIds([]int64{id1, id2})
	if err != nil {
		t.Fatalf("FindByIds() error = %v", err)
	}
	if len(subset) != 2 {
		t.Errorf("FindByIds() len = %d; want 2", len(subset))
	}
}

func TestEmployeeRepository_DeleteById(t *testing.T) {
	TruncateTable(testDB)

	repo := employee.NewEmployeeRepository(testDB)
	now := time.Now()
	id, _ := repo.Save(&employee.EmployeeEntity{Name: "Frank", CreatedAt: now, UpdatedAt: now})

	if err := repo.DeleteById(id); err != nil {
		t.Fatalf("DeleteById(%d) error = %v", id, err)
	}

	all, _ := repo.FindAll()
	if len(all) != 0 {
		t.Errorf("After DeleteById, FindAll() len = %d; want 0", len(all))
	}
}

func TestEmployeeRepository_DeleteByIds(t *testing.T) {
	TruncateTable(testDB)

	repo := employee.NewEmployeeRepository(testDB)
	now := time.Now()
	id1, _ := repo.Save(&employee.EmployeeEntity{Name: "George", CreatedAt: now, UpdatedAt: now})
	id2, _ := repo.Save(&employee.EmployeeEntity{Name: "Hannah", CreatedAt: now, UpdatedAt: now})

	if err := repo.DeleteByIds([]int64{id1, id2}); err != nil {
		t.Fatalf("DeleteByIds(%v) error = %v", []int64{id1, id2}, err)
	}
	all, _ := repo.FindAll()
	if len(all) != 0 {
		t.Errorf("After DeleteByIds, FindAll() len = %d; want 0", len(all))
	}
}
