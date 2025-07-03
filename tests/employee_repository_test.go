package tests

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"idm/inner/employee"
	"testing"
	"time"
)

func TruncateTable(db *sqlx.DB) {
	_, err := db.Exec("TRUNCATE employee RESTART IDENTITY CASCADE")
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

// Test BeginTransaction and Rollback: changes should not persist after rollback
func TestRepository_BeginTransaction_Rollback(t *testing.T) {
	TruncateTable(testDB)
	repo := employee.NewEmployeeRepository(testDB)
	// Begin a transaction
	tx, err := repo.BeginTransaction()
	if err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	// Insert a record within the transaction
	_, err = tx.Exec("INSERT INTO employee (name, created_at, updated_at) VALUES ($1, now(), now())", "Temp")
	if err != nil {
		t.Fatalf("tx.Exec insert error = %v", err)
	}
	// Rollback
	if err := tx.Rollback(); err != nil {
		t.Fatalf("tx.Rollback() error = %v", err)
	}
	// After rollback, the record should not exist
	var cnt int
	if err := testDB.Get(&cnt, "SELECT COUNT(1) FROM employee WHERE name = $1", "Temp"); err != nil {
		t.Fatalf("Get count error = %v", err)
	}
	if cnt != 0 {
		t.Errorf("after rollback, expected 0 rows, got %d", cnt)
	}
}

// Test FindByNameTx within a transaction
func TestRepository_FindByNameTx(t *testing.T) {
	TruncateTable(testDB)
	repo := employee.NewEmployeeRepository(testDB)
	tx, err := repo.BeginTransaction()
	if err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	// Initially, no such name
	exists, err := repo.FindByNameTx(tx, "Nobody")
	if err != nil {
		t.Fatalf("FindByNameTx() error = %v", err)
	}
	if exists {
		t.Errorf("expected exists=false for name 'Nobody', got true")
	}
	// Insert a row within the same transaction
	_, err = tx.Exec("INSERT INTO employee (name, created_at, updated_at) VALUES ($1, now(), now())", "Alice")
	if err != nil {
		t.Fatalf("tx.Exec insert error = %v", err)
	}
	// Now FindByNameTx should see it
	exists, err = repo.FindByNameTx(tx, "Alice")
	if err != nil {
		t.Fatalf("FindByNameTx() error after insert = %v", err)
	}
	if !exists {
		t.Errorf("expected exists=true for name 'Alice', got false")
	}
}

// Test SaveTx: insert via transaction and commit persists record
func TestRepository_SaveTx(t *testing.T) {
	TruncateTable(testDB)
	repo := employee.NewEmployeeRepository(testDB)
	tx, err := repo.BeginTransaction()
	if err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	// Use SaveTx to insert a new entity
	id, err := repo.SaveTx(tx, employee.Entity{Name: "Bob"})
	if err != nil {
		t.Fatalf("SaveTx() error = %v", err)
	}
	// Commit the transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("tx.Commit() error = %v", err)
	}
	// Verify record exists
	var got employee.Entity
	if err := testDB.Get(&got, "SELECT * FROM employee WHERE id = $1", id); err != nil {
		t.Fatalf("FindById via testDB error = %v", err)
	}
	if got.Id != id || got.Name != "Bob" {
		t.Errorf("expected saved entity {Id:%d Name:'Bob'}, got %+v", id, got)
	}
}
