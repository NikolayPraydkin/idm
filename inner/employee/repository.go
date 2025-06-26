package employee

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewEmployeeRepository(database *sqlx.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) FindById(id int64) (*Entity, error) {
	var entity Entity
	err := r.db.Get(&entity, "SELECT * FROM employee WHERE id = $1", id)
	return &entity, err
}

func (r *Repository) Add(employee *Entity) error {
	_, err := r.db.NamedExec(`INSERT INTO employee (name, created_at, updated_at) 
		VALUES (:name, :created_at, :updated_at)`, employee)
	return err
}

func (r *Repository) Save(employee *Entity) (int64, error) {
	var id int64
	query := `INSERT INTO employee (name, created_at, updated_at)
			  VALUES (:name, :created_at, :updated_at)
			  RETURNING id`
	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return 0, err
	}
	err = stmt.QueryRowx(employee).Scan(&id)
	return id, err
}

func (r *Repository) FindAll() ([]Entity, error) {
	var employees []Entity
	err := r.db.Select(&employees, "SELECT * FROM employee")
	return employees, err
}

func (r *Repository) FindByIds(ids []int64) ([]Entity, error) {
	query, args, err := sqlx.In("SELECT * FROM employee WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	var employees []Entity
	err = r.db.Select(&employees, query, args...)
	return employees, err
}

func (r *Repository) DeleteById(id int64) error {
	_, err := r.db.Exec("DELETE FROM employee WHERE id = $1", id)
	return err
}

func (r *Repository) DeleteByIds(ids []int64) error {
	query, args, err := sqlx.In("DELETE FROM employee WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.Exec(query, args...)
	return err
}

func (r *Repository) TransactionalCreate(e *Entity) (idReturn int64, err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			errTx := tx.Rollback()
			if errTx != nil {
				err = fmt.Errorf("creating employee: rolling back transaction errors: %w, %w", err, errTx)
			}
		}
	}()

	// проверяем, есть ли в БД сотрудник с таким именем
	var cnt int
	if err := tx.Get(&cnt, "SELECT COUNT(1) FROM employee WHERE name = $1", e.Name); err != nil {
		err := tx.Rollback()
		if err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("error checking existing employee: %w", err)
	}
	if cnt > 0 {
		err := tx.Rollback()
		if err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("employee with name %q already exists", e.Name)
	}

	var id int64
	stmt, err := tx.PrepareNamed(`
        INSERT INTO employee (name, created_at, updated_at)
        VALUES (:name, :created_at, :updated_at)
        RETURNING id
    `)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("failed to prepare insert: %w", err)
	}
	if err := stmt.QueryRowx(e).Scan(&id); err != nil {
		err := tx.Rollback()
		if err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("error inserting employee: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return id, nil
}
