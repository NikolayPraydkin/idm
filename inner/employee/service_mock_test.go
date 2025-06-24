package employee

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

// --- 1. Определяем мок-репозиторий ---
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) FindById(id int64) (*Entity, error) {
	args := m.Called(id)
	// Может быть nil, поэтому проверяем
	if ent, ok := args.Get(0).(*Entity); ok {
		return ent, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepo) Add(e *Entity) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockRepo) Save(e *Entity) (int64, error) {
	args := m.Called(e)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepo) FindAll() ([]Entity, error) {
	args := m.Called()
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRepo) FindByIds(ids []int64) ([]Entity, error) {
	args := m.Called(ids)
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRepo) DeleteById(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepo) DeleteByIds(ids []int64) error {
	args := m.Called(ids)
	return args.Error(0)
}

// --- 2. Сам сервис для тестов ---
func newTestService(repo Repo) *Service {
	return NewService(repo)
}

// --- 3. Тесты ---
func TestService_FindById(t *testing.T) {
	repo := new(MockRepo)
	svc := newTestService(repo)

	now := time.Now()
	ent := &Entity{Id: 1, Name: "John", CreatedAt: now, UpdatedAt: now}
	resp := Response{Id: 1, Name: "John", CreatedAt: now, UpdatedAt: now}

	// Успешный кейс
	repo.On("FindById", int64(1)).Return(ent, nil)
	got, err := svc.FindById(1)
	assert.NoError(t, err)
	assert.Equal(t, resp, got)
	repo.AssertCalled(t, "FindById", int64(1))

	// Ошибка из репо
	repo = new(MockRepo)
	svc = newTestService(repo)
	repo.On("FindById", int64(2)).Return((*Entity)(nil), errors.New("not found"))
	_, err = svc.FindById(2)
	assert.Error(t, err)
	repo.AssertNumberOfCalls(t, "FindById", 1)
}

func TestService_Add(t *testing.T) {
	repo := new(MockRepo)
	svc := newTestService(repo)

	ent := Entity{Id: 0, Name: "Jane"}
	repo.On("Add", &ent).Return(nil)

	err := svc.Add(ent)
	assert.NoError(t, err)
	repo.AssertCalled(t, "Add", &ent)
}

func TestService_Save(t *testing.T) {
	repo := new(MockRepo)
	svc := newTestService(repo)

	ent := Entity{Id: 0, Name: "Bob"}
	repo.On("Save", &ent).Return(int64(42), nil)

	id, err := svc.Save(ent)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), id)
	repo.AssertCalled(t, "Save", &ent)
}

func TestService_FindAll(t *testing.T) {
	repo := new(MockRepo)
	svc := newTestService(repo)

	now := time.Now()
	ents := []Entity{
		{Id: 1, Name: "A", CreatedAt: now, UpdatedAt: now},
		{Id: 2, Name: "B", CreatedAt: now, UpdatedAt: now},
	}
	repo.On("FindAll").Return(ents, nil)

	got, err := svc.FindAll()
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, ents[0].toResponse(), got[0])
	assert.Equal(t, ents[1].toResponse(), got[1])
	repo.AssertNumberOfCalls(t, "FindAll", 1)
}

func TestService_FindByIds(t *testing.T) {
	repo := new(MockRepo)
	svc := newTestService(repo)

	now := time.Now()
	ids := []int64{1, 3}
	ents := []Entity{
		{Id: 1, Name: "X", CreatedAt: now, UpdatedAt: now},
		{Id: 3, Name: "Y", CreatedAt: now, UpdatedAt: now},
	}
	repo.On("FindByIds", ids).Return(ents, nil)

	got, err := svc.FindByIds(ids)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	repo.AssertCalled(t, "FindByIds", ids)
}

func TestService_DeleteById(t *testing.T) {
	repo := new(MockRepo)
	svc := newTestService(repo)

	repo.On("DeleteById", int64(5)).Return(nil)
	err := svc.DeleteById(5)
	assert.NoError(t, err)
	repo.AssertNumberOfCalls(t, "DeleteById", 1)
}

func TestService_DeleteByIds(t *testing.T) {
	repo := new(MockRepo)
	svc := newTestService(repo)

	ids := []int64{7, 8}
	repo.On("DeleteByIds", ids).Return(nil)
	err := svc.DeleteByIds(ids)
	assert.NoError(t, err)
	repo.AssertCalled(t, "DeleteByIds", ids)
}
