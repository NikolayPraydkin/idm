package employee

import "time"

type Entity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (e *Entity) toResponse() Response {
	return Response{
		Id:        e.Id,
		Name:      e.Name,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

type Response struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRequest struct {
	Name string `json:"name" validate:"required,min=2,max=155"`
}

func (req *CreateRequest) ToEntity() *Entity {
	return &Entity{Name: req.Name}
}

type PageRequest struct {
	PageSize   int `validate:"min=1,max=100"`
	PageNumber int `validate:"min=0"`
	TextFilter string
}

type PageResponse struct {
	Result     any   `json:"result"`
	PageSize   int   `json:"page_size"`
	PageNumber int   `json:"page_number"`
	Total      int64 `json:"total"`
}
