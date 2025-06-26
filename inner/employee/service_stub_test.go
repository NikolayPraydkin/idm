package employee

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// stub-репозиторий, реализующий только FindAll
type StubRepo struct{}

func (s *StubRepo) TransactionalCreate(e *Entity) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubRepo) FindById(id int64) (*Entity, error) {
	panic("not implemented")
}
func (s *StubRepo) Add(e *Entity) error {
	panic("not implemented")
}
func (s *StubRepo) Save(e *Entity) (int64, error) {
	panic("not implemented")
}
func (s *StubRepo) FindAll() ([]Entity, error) {
	// жёстко зашитые данные
	now := time.Date(2025, 6, 24, 12, 0, 0, 0, time.UTC)
	return []Entity{
		{Id: 10, Name: "Stub A", CreatedAt: now, UpdatedAt: now},
		{Id: 20, Name: "Stub B", CreatedAt: now, UpdatedAt: now},
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

func TestFindAll_WithStub(t *testing.T) {
	svc := NewService(&StubRepo{})

	resps, err := svc.FindAll()
	assert.NoError(t, err)
	// Должны получить ровно две записи из stub
	assert.Len(t, resps, 2)
	assert.Equal(t, int64(10), resps[0].Id)
	assert.Equal(t, "Stub A", resps[0].Name)
	assert.Equal(t, int64(20), resps[1].Id)
	assert.Equal(t, "Stub B", resps[1].Name)
}
