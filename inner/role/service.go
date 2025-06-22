package role

import (
	"fmt"
)

type Service struct {
	repo Repo
}

type Repo interface {
	Add(e *Entity) error
	FindById(id int64) (*Entity, error)
	FindAll() ([]Entity, error)
	FindByIds(ids []int64) ([]Entity, error)
	DeleteById(id int64) error
	DeleteByIds(ids []int64) error
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (svc *Service) Add(e Entity) error {
	return svc.repo.Add(&e)
}

func (svc *Service) FindById(id int64) (Response, error) {
	e, err := svc.repo.FindById(id)
	if err != nil {
		return Response{}, fmt.Errorf("failed to find role with id %d: %w", id, err)
	}
	return e.toResponse(), nil
}

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

func (svc *Service) DeleteById(id int64) error {
	return svc.repo.DeleteById(id)
}

func (svc *Service) DeleteByIds(ids []int64) error {
	return svc.repo.DeleteByIds(ids)
}
