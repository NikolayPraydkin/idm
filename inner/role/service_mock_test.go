package role

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---- 1. Мок-репозиторий ----
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Add(e *Entity) error {
	return m.Called(e).Error(0)
}

func (m *MockRepo) FindById(id int64) (*Entity, error) {
	args := m.Called(id)
	if ent, ok := args.Get(0).(*Entity); ok {
		return ent, args.Error(1)
	}
	return nil, args.Error(1)
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
	return m.Called(id).Error(0)
}

func (m *MockRepo) DeleteByIds(ids []int64) error {
	return m.Called(ids).Error(0)
}

// ---- 2. Тесты для Service ----
func TestService_AllMethods_WithMock(t *testing.T) {
	now := time.Now()
	// Пример Entity
	ent1 := &Entity{Id: 1, Name: "Admin", CreatedAt: now, UpdatedAt: now}
	ent2 := &Entity{Id: 2, Name: "User", CreatedAt: now, UpdatedAt: now}
	ents := []Entity{*ent1, *ent2}

	repo := new(MockRepo)
	svc := NewService(repo)

	t.Run("Add calls repo.Add", func(t *testing.T) {
		e := Entity{Id: 0, Name: "Guest"}
		repo.On("Add", &e).Return(nil)
		err := svc.Add(e)
		assert.NoError(t, err)
		repo.AssertCalled(t, "Add", &e)
	})

	t.Run("FindById success", func(t *testing.T) {
		repo.On("FindById", int64(1)).Return(ent1, nil)
		resp, err := svc.FindById(1)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), resp.Id)
		assert.Equal(t, "Admin", resp.Name)
		repo.AssertCalled(t, "FindById", int64(1))
	})

	t.Run("FindById error", func(t *testing.T) {
		repoErr := errors.New("not found")
		repo.On("FindById", int64(99)).Return((*Entity)(nil), repoErr)
		_, err := svc.FindById(99)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find role")
		repo.AssertCalled(t, "FindById", int64(99))
	})

	t.Run("FindAll returns all", func(t *testing.T) {
		repo.On("FindAll").Return(ents, nil)
		resps, err := svc.FindAll()
		assert.NoError(t, err)
		assert.Len(t, resps, 2)
		assert.Equal(t, int64(1), resps[0].Id)
		assert.Equal(t, int64(2), resps[1].Id)
		repo.AssertNumberOfCalls(t, "FindAll", 1)
	})

	t.Run("FindByIds filters", func(t *testing.T) {
		ids := []int64{1, 2}
		repo.On("FindByIds", ids).Return(ents, nil)
		resps, err := svc.FindByIds(ids)
		assert.NoError(t, err)
		assert.Len(t, resps, 2)
		repo.AssertCalled(t, "FindByIds", ids)
	})

	t.Run("DeleteById calls repo", func(t *testing.T) {
		repo.On("DeleteById", int64(2)).Return(nil)
		err := svc.DeleteById(2)
		assert.NoError(t, err)
		repo.AssertCalled(t, "DeleteById", int64(2))
	})

	t.Run("DeleteByIds calls repo", func(t *testing.T) {
		ids := []int64{1, 2}
		repo.On("DeleteByIds", ids).Return(nil)
		err := svc.DeleteByIds(ids)
		assert.NoError(t, err)
		repo.AssertCalled(t, "DeleteByIds", ids)
	})
}
