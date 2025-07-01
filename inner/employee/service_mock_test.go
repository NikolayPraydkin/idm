package employee

import (
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"regexp"
	"testing"
	"time"
)

// --- 1. Определяем мок-репозиторий ---
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) BeginTransaction() (*sqlx.Tx, error) {
	args := m.Called()
	return args.Get(0).(*sqlx.Tx), args.Error(1)
}

func (m *MockRepo) FindByNameTx(tx *sqlx.Tx, name string) (bool, error) {
	args := m.Called(tx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepo) SaveTx(tx *sqlx.Tx, employee Entity) (int64, error) {
	args := m.Called(tx, employee)
	return args.Get(0).(int64), args.Error(1)
}

// ErrEmployeeAlreadyExists возвращается, если работник с таким именем уже существует
var ErrEmployeeAlreadyExists = fmt.Errorf("employee already exists")

func (m *MockRepo) TransactionalCreate(e *Entity) (int64, error) {
	args := m.Called(e)
	// Optionally mutate e in success cases via Run in tests
	return args.Get(0).(int64), args.Error(1)
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

func TestService_SaveWithTransaction(t *testing.T) {
	type testCase struct {
		name   string
		setup  func()
		verify func(*testing.T)
	}

	tests := []testCase{
		{
			name: "begin transaction error",
			setup: func() {
			},
			verify: func(t *testing.T) {
			},
		},
		{
			name: "check existence error",
			setup: func() {
			},
			verify: func(t *testing.T) {
			},
		},
		{
			name: "duplicate employee",
			setup: func() {
			},
			verify: func(t *testing.T) {
			},
		},
		{
			name: "insert error",
			setup: func() {
			},
			verify: func(t *testing.T) {
			},
		},
		{
			name: "success creation",
			setup: func() {
			},
			verify: func(t *testing.T) {
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// create fresh mock for this subtest
			dbMock, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer dbMock.Close()

			// wrap sql.DB into sqlx.DB and create new svc
			db := sqlx.NewDb(dbMock, "sqlmock")
			repo := NewEmployeeRepository(db)
			svc := NewService(repo)

			// common entity
			entity := Entity{Name: "Alice"}

			// redefine setup and verify closures to capture mock, svc, and entity
			tc.setup = func() {
				switch tc.name {
				case "begin transaction error":
					mock.ExpectBegin().WillReturnError(errors.New("begin failed"))
				case "check existence error":
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta("select exists(select 1 from employee where name = $1)")).
						WithArgs(entity.Name).
						WillReturnError(errors.New("select failed"))
					mock.ExpectRollback()
				case "duplicate employee":
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta("select exists(select 1 from employee where name = $1)")).
						WithArgs(entity.Name).
						WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
					mock.ExpectRollback()
				case "insert error":
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta("select exists(select 1 from employee where name = $1)")).
						WithArgs(entity.Name).
						WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
					mock.ExpectQuery(regexp.QuoteMeta("insert into employee (name) values ($1) returning id")).
						WithArgs(entity.Name).
						WillReturnError(errors.New("insert failed"))
					mock.ExpectRollback()
				case "success creation":
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta("select exists(select 1 from employee where name = $1)")).
						WithArgs(entity.Name).
						WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
					mock.ExpectQuery(regexp.QuoteMeta("insert into employee (name) values ($1) returning id")).
						WithArgs(entity.Name).
						WillReturnRows(sqlmock.NewRows([]string{"employeeid"}).AddRow(123))
					mock.ExpectCommit()
				}
			}

			tc.verify = func(t *testing.T) {
				id, err := svc.SaveWithTransaction(entity)
				switch tc.name {
				case "begin transaction error":
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "error creating transaction")
				case "check existence error":
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "error finding employee by name")
				case "duplicate employee":
					assert.Error(t, err)
					assert.Contains(t, err.Error(), ErrEmployeeAlreadyExists.Error())
				case "insert error":
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "error creating employee with name")
				case "success creation":
					assert.NoError(t, err)
					assert.Equal(t, int64(123), id)
				}
			}

			tc.setup()
			tc.verify(t)

			// ensure expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
