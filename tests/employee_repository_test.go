package tests

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"idm/inner/employee"
	"testing"
	"time"
)

func TruncateTable(db *sqlx.DB) {
	_, err := testDB.Exec("TRUNCATE employee RESTART IDENTITY CASCADE")
	if err != nil {
		panic(fmt.Errorf("failed TRUNCATE role: %v", err))
	}
}

func TestEmployeeRepository_AddAndFindById(t *testing.T) {
	TruncateTable(testDB)

	repo := employee.NewEmployeeRepository(testDB)
	now := time.Now()
	e := &employee.Entity{
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
	_, _ = repo.Save(&employee.Entity{Name: "Bob", CreatedAt: now, UpdatedAt: now})
	_, _ = repo.Save(&employee.Entity{Name: "Carol", CreatedAt: now, UpdatedAt: now})

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
	id1, _ := repo.Save(&employee.Entity{Name: "Dave", CreatedAt: now, UpdatedAt: now})
	id2, _ := repo.Save(&employee.Entity{Name: "Eve", CreatedAt: now, UpdatedAt: now})

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
	id, _ := repo.Save(&employee.Entity{Name: "Frank", CreatedAt: now, UpdatedAt: now})

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
	id1, _ := repo.Save(&employee.Entity{Name: "George", CreatedAt: now, UpdatedAt: now})
	id2, _ := repo.Save(&employee.Entity{Name: "Hannah", CreatedAt: now, UpdatedAt: now})

	if err := repo.DeleteByIds([]int64{id1, id2}); err != nil {
		t.Fatalf("DeleteByIds(%v) error = %v", []int64{id1, id2}, err)
	}
	all, _ := repo.FindAll()
	if len(all) != 0 {
		t.Errorf("After DeleteByIds, FindAll() len = %d; want 0", len(all))
	}
}

func TestTransactionalCreate_SuccessAndDuplicate(t *testing.T) {
	TruncateTable(testDB)
	repo := employee.NewEmployeeRepository(testDB)

	e := &employee.Entity{
		Name:      "Alice",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := repo.TransactionalCreate(e)
	fmt.Printf("%v", id)
	assert.NoError(t, err)
	var cnt int
	err = testDB.Get(&cnt, "SELECT COUNT(1) FROM employee WHERE name = $1", "Alice")
	if err != nil {
		t.Logf("duplicate insert error: %v", err)
	}
	assert.Equal(t, 1, cnt)

	// Повторная вставка
	e2 := &employee.Entity{
		Name:      "Alice",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = repo.TransactionalCreate(e2)
	assert.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("employee with name %q already exists", e2.Name))
	testDB.Get(&cnt, "SELECT COUNT(1) FROM employee WHERE name = $1", "Alice")
	assert.Equal(t, 1, cnt)
}

func TestTransactionalCreate_InsertErrorRollsBack(t *testing.T) {
	TruncateTable(testDB)
	repo := employee.NewEmployeeRepository(testDB)
	e := &employee.Entity{}
	_, err := repo.TransactionalCreate(e)
	assert.Error(t, err)
	var cnt int
	testDB.Get(&cnt, "SELECT COUNT(1) FROM employee")
	assert.Equal(t, 0, cnt)
}
