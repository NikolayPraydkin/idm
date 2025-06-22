package role

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(database *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: database}
}

type RoleEntity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *RoleRepository) Add(role *RoleEntity) error {
	_, err := r.db.NamedExec(`INSERT INTO role (name, created_at, updated_at) 
		VALUES (:name, :created_at, :updated_at)`, role)
	return err
}

func (r *RoleRepository) FindById(id int64) (*RoleEntity, error) {
	var entity RoleEntity
	err := r.db.Get(&entity, "SELECT * FROM role WHERE id = $1", id)
	return &entity, err
}

func (r *RoleRepository) FindAll() (roles []RoleEntity, err error) {
	err = r.db.Select(&roles, "SELECT * FROM role")
	return roles, err
}

func (r *RoleRepository) FindByIds(ids []int64) ([]RoleEntity, error) {
	query, args, err := sqlx.In("SELECT * FROM role WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	var roles []RoleEntity
	err = r.db.Select(&roles, query, args...)
	return roles, err
}

func (r *RoleRepository) DeleteById(id int64) error {
	_, err := r.db.Exec("DELETE FROM role WHERE id = $1", id)
	return err
}

func (r *RoleRepository) DeleteByIds(ids []int64) error {
	query, args, err := sqlx.In("DELETE FROM role WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.Exec(query, args...)
	return err
}
