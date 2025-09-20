package employee

import (
	"context"
	"fmt"
	"idm/inner/common"
	"idm/inner/validator"

	"github.com/jmoiron/sqlx"
)

// структура Service, которая будет инкапсулировать бизнес-логику
type Service struct {
	repo      Repo
	validator *validator.Validator
}

// интерфейс репозитория
// определяет, какие методы требуются от реализации репозитория
// (здесь все из employee.Repository)
type Repo interface {
	FindById(id int64) (*Entity, error)
	Add(e *Entity) error
	Save(e *Entity) (int64, error)
	FindAll() ([]Entity, error)
	FindByIds(ids []int64) ([]Entity, error)
	DeleteById(id int64) error
	DeleteByIds(ids []int64) error
	BeginTransaction() (*sqlx.Tx, error)
	FindByNameTx(tx *sqlx.Tx, name string) (bool, error)
	SaveTx(tx *sqlx.Tx, employee *Entity) (int64, error)
	FindEmployeesPage(req PageRequest) ([]Entity, int64, error)
}

// функция-конструктор
func NewService(repo Repo) *Service {
	return &Service{repo: repo, validator: validator.New()}
}

// бизнес-логика получения одного работника по id
func (svc *Service) FindById(ctx context.Context, id int64) (Response, error) {
	employee, err := svc.repo.FindById(id)
	if err != nil {
		return Response{}, fmt.Errorf("error finding employee with id %d: %w", id, err)
	}
	return employee.toResponse(), nil
}

// добавление работника без возврата id
func (svc *Service) Add(ctx context.Context, req CreateRequest) error {
	if err := svc.validator.Validate(req); err != nil {
		return common.RequestValidationError{Message: err.Error()}
	}
	return svc.repo.Add(req.ToEntity())
}

// добавление с возвратом id
func (svc *Service) Save(ctx context.Context, req CreateRequest) (int64, error) {
	if err := svc.validator.Validate(req); err != nil {
		return 0, common.RequestValidationError{Message: err.Error()}
	}
	return svc.repo.Save(req.ToEntity())
}

// получить всех работников
func (svc *Service) FindAll(ctx context.Context) ([]Response, error) {
	entities, err := svc.repo.FindAll()
	if err != nil {
		return nil, err
	}
	var result []Response
	for _, e := range entities {
		result = append(result, e.toResponse())
	}
	return result, nil
}

// получить работников по слайсу id
func (svc *Service) FindByIds(ctx context.Context, ids []int64) ([]Response, error) {
	entities, err := svc.repo.FindByIds(ids)
	if err != nil {
		return nil, err
	}
	var result []Response
	for _, e := range entities {
		result = append(result, e.toResponse())
	}
	return result, nil
}

// удалить одного по id
func (svc *Service) DeleteById(ctx context.Context, id int64) error {
	return svc.repo.DeleteById(id)
}

// удалить всех по слайсу id
func (svc *Service) DeleteByIds(ctx context.Context, ids []int64) error {
	return svc.repo.DeleteByIds(ids)
}

// SaveWithTransaction проверяет дубликаты и создаёт запись в рамках одной транзакции.
func (svc *Service) SaveWithTransaction(ctx context.Context, e CreateRequest) (int64, error) {

	if err := svc.validator.Validate(e); err != nil {
		return 0, common.RequestValidationError{Message: err.Error()}
	}
	tx, err := svc.repo.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("error creating transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("creating employee panic: %v", r)
			errTx := tx.Rollback()
			if errTx != nil {
				err = fmt.Errorf("creating employee: rolling back transaction errors: %w, %w", err, errTx)
			}
		} else if err != nil {
			errTx := tx.Rollback()
			if errTx != nil {
				err = fmt.Errorf("creating employee: rolling back transaction errors: %w, %w", err, errTx)
			}
		} else {
			errTx := tx.Commit()
			if errTx != nil {
				err = fmt.Errorf("creating employee: commiting transaction error: %w", errTx)
			}
		}
	}()
	if err != nil {
		return 0, fmt.Errorf("error create employee: error creating transaction: %w", err)
	}
	isExist, err := svc.repo.FindByNameTx(tx, e.Name)
	if err != nil {
		return 0, fmt.Errorf("error finding employee by name: %s, %w", e.Name, err)
	}
	if isExist {
		err = common.AlreadyExistsError{Message: "employee already exists"}
		return 0, err
	}
	newEmployeeId, err := svc.repo.SaveTx(tx, e.ToEntity())
	if err != nil {
		err = fmt.Errorf("error creating employee with name: %s %v", e.Name, err)
	}
	return newEmployeeId, err
}

func (svc *Service) GetEmployeesPage(ctx context.Context, req PageRequest) (PageResponse, error) {
	if err := svc.validator.Validate(req); err != nil {
		return PageResponse{}, common.RequestValidationError{Message: err.Error()}
	}
	entities, total, err := svc.repo.FindEmployeesPage(req)
	if err != nil {
		return PageResponse{}, err
	}

	var respItems []Response
	for _, e := range entities {
		respItems = append(respItems, e.toResponse())
	}

	return PageResponse{
		Result:     respItems,
		PageSize:   req.PageSize,
		PageNumber: req.PageNumber,
		Total:      total,
	}, nil
}
