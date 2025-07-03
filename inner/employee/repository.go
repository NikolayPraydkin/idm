package employee

import (
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

func (r *Repository) BeginTransaction() (tx *sqlx.Tx, err error) {
	return r.db.Beginx()
}

func (r *Repository) FindByNameTx(tx *sqlx.Tx, name string) (isExists bool, err error) {
	err = tx.Get(
		&isExists,
		"select exists(select 1 from employee where name = $1)",
		name,
	)
	return isExists, err
}

func (r *Repository) SaveTx(tx *sqlx.Tx, employee Entity) (employeeId int64, err error) {
	err = tx.Get(
		&employeeId,
		`insert into employee (name) values ($1) returning id`,
		employee.Name,
	)
	return employeeId, err
}
