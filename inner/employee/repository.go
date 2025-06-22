package employee

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type EmployeeRepository struct {
	db *sqlx.DB
}

func NewEmployeeRepository(database *sqlx.DB) *EmployeeRepository {
	return &EmployeeRepository{db: database}
}

type EmployeeEntity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *EmployeeRepository) FindById(id int64) (*EmployeeEntity, error) {
	var entity EmployeeEntity
	err := r.db.Get(&entity, "SELECT * FROM employee WHERE id = $1", id)
	return &entity, err
}

func (r *EmployeeRepository) Add(employee *EmployeeEntity) error {
	_, err := r.db.NamedExec(`INSERT INTO employee (name, created_at, updated_at) 
		VALUES (:name, :created_at, :updated_at)`, employee)
	return err
}

func (r *EmployeeRepository) Save(employee *EmployeeEntity) (int64, error) {
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

func (r *EmployeeRepository) FindAll() ([]EmployeeEntity, error) {
	var employees []EmployeeEntity
	err := r.db.Select(&employees, "SELECT * FROM employee")
	return employees, err
}

func (r *EmployeeRepository) FindByIds(ids []int64) ([]EmployeeEntity, error) {
	query, args, err := sqlx.In("SELECT * FROM employee WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	var employees []EmployeeEntity
	err = r.db.Select(&employees, query, args...)
	return employees, err
}

func (r *EmployeeRepository) DeleteById(id int64) error {
	_, err := r.db.Exec("DELETE FROM employee WHERE id = $1", id)
	return err
}

func (r *EmployeeRepository) DeleteByIds(ids []int64) error {
	query, args, err := sqlx.In("DELETE FROM employee WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.Exec(query, args...)
	return err
}
