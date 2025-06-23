package tests

import (
	"fmt"
	"idm/inner/role"
	"testing"
	"time"
)

func TruncateRoleTable() {
	_, err := testDB.Exec("TRUNCATE role RESTART IDENTITY CASCADE")
	if err != nil {
		panic(fmt.Errorf("failed TRUNCATE role: %v", err))
	}
}

func TestRoleRepository_AddAndFindById(t *testing.T) {
	TruncateRoleTable()

	repo := role.NewRoleRepository(testDB)
	now := time.Now()
	r := &role.Entity{
		Name:      "Manager",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := repo.Add(r)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	roles, _ := repo.FindAll()
	if len(roles) != 1 {
		t.Fatalf("Expected 1 role, got %d", len(roles))
	}

	got, err := repo.FindById(roles[0].Id)
	if err != nil {
		t.Fatalf("FindById() error = %v", err)
	}
	if got.Name != "Manager" {
		t.Errorf("FindById().Name = %q; want %q", got.Name, "Manager")
	}
}

func TestRoleRepository_FindAll(t *testing.T) {
	TruncateRoleTable()
	repo := role.NewRoleRepository(testDB)
	now := time.Now()
	err := repo.Add(&role.Entity{Name: "Dev", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	err = repo.Add(&role.Entity{Name: "QA", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	all, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(all) != 2 {
		t.Errorf("FindAll() len = %d; want 2", len(all))
	}
}

func TestRoleRepository_FindByIds(t *testing.T) {
	TruncateRoleTable()
	repo := role.NewRoleRepository(testDB)
	now := time.Now()
	err := repo.Add(&role.Entity{Name: "PM", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	err = repo.Add(&role.Entity{Name: "Support", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	all, _ := repo.FindAll()
	ids := []int64{all[0].Id, all[1].Id}

	result, err := repo.FindByIds(ids)
	if err != nil {
		t.Fatalf("FindByIds() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("FindByIds() len = %d; want 2", len(result))
	}
}

func TestRoleRepository_DeleteById(t *testing.T) {
	TruncateRoleTable()
	repo := role.NewRoleRepository(testDB)
	now := time.Now()
	err := repo.Add(&role.Entity{Name: "Temp", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	all, _ := repo.FindAll()
	id := all[0].Id

	err = repo.DeleteById(id)
	if err != nil {
		t.Fatalf("DeleteById() error = %v", err)
	}
	remaining, _ := repo.FindAll()
	if len(remaining) != 0 {
		t.Errorf("After DeleteById, FindAll() len = %d; want 0", len(remaining))
	}
}

func TestRoleRepository_DeleteByIds(t *testing.T) {
	TruncateRoleTable()
	repo := role.NewRoleRepository(testDB)
	now := time.Now()
	err := repo.Add(&role.Entity{Name: "Intern", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	err = repo.Add(&role.Entity{Name: "Contractor", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	all, _ := repo.FindAll()
	ids := []int64{all[0].Id, all[1].Id}

	err = repo.DeleteByIds(ids)
	if err != nil {
		t.Fatalf("DeleteByIds() error = %v", err)
	}
	remaining, _ := repo.FindAll()
	if len(remaining) != 0 {
		t.Errorf("After DeleteByIds, FindAll() len = %d; want 0", len(remaining))
	}
}
