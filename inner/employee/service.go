package employee

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

// структура Service, которая будет инкапсулировать бизнес-логику
type Service struct {
	repo Repo
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
	SaveTx(tx *sqlx.Tx, employee Entity) (int64, error)
}

// функция-конструктор
func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

// бизнес-логика получения одного работника по id
func (svc *Service) FindById(id int64) (Response, error) {
	employee, err := svc.repo.FindById(id)
	if err != nil {
		return Response{}, fmt.Errorf("error finding employee with id %d: %w", id, err)
	}
	return employee.toResponse(), nil
}

// добавление работника без возврата id
func (svc *Service) Add(req Entity) error {
	return svc.repo.Add(&req)
}

// добавление с возвратом id
func (svc *Service) Save(req Entity) (int64, error) {
	return svc.repo.Save(&req)
}

// получить всех работников
func (svc *Service) FindAll() ([]Response, error) {
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
func (svc *Service) FindByIds(ids []int64) ([]Response, error) {
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
func (svc *Service) DeleteById(id int64) error {
	return svc.repo.DeleteById(id)
}

// удалить всех по слайсу id
func (svc *Service) DeleteByIds(ids []int64) error {
	return svc.repo.DeleteByIds(ids)
}

// SaveWithTransaction проверяет дубликаты и создаёт запись в рамках одной транзакции.
func (svc *Service) SaveWithTransaction(e Entity) (int64, error) {
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
		err = fmt.Errorf("employee already exists")
		return 0, err
	}
	newEmployeeId, err := svc.repo.SaveTx(tx, e)
	if err != nil {
		err = fmt.Errorf("error creating employee with name: %s %v", e.Name, err)
	}
	return newEmployeeId, err
}
