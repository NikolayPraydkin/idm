package employee

import (
	"fmt"
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
