package role

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ---- StubRepo ----
type StubRepo struct{}

func (s *StubRepo) Add(e *Entity) error {
	panic("not implemented")
}
func (s *StubRepo) FindById(id int64) (*Entity, error) {
	panic("not implemented")
}
func (s *StubRepo) FindAll() ([]Entity, error) {
	// Жёстко зашитые данные, никакого testify
	now := time.Date(2025, 6, 24, 15, 0, 0, 0, time.UTC)
	return []Entity{
		{Id: 10, Name: "StubRoleA", CreatedAt: now, UpdatedAt: now},
		{Id: 20, Name: "StubRoleB", CreatedAt: now, UpdatedAt: now},
	}, nil
}
func (s *StubRepo) FindByIds(ids []int64) ([]Entity, error) {
	panic("not implemented")
}
func (s *StubRepo) DeleteById(id int64) error {
	panic("not implemented")
}
func (s *StubRepo) DeleteByIds(ids []int64) error {
	panic("not implemented")
}

// ---- Тест через stub ----
func Test_FindAll_WithStub(t *testing.T) {
	svc := NewService(&StubRepo{})

	resps, err := svc.FindAll()
	assert.NoError(t, err)
	assert.Len(t, resps, 2)

	// Проверяем, что StubRepo действительно вернул то, что нам нужно
	assert.Equal(t, int64(10), resps[0].Id)
	assert.Equal(t, "StubRoleA", resps[0].Name)
	assert.Equal(t, int64(20), resps[1].Id)
	assert.Equal(t, "StubRoleB", resps[1].Name)
}
