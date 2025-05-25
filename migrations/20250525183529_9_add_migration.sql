-- +goose Up
-- +goose StatementBegin
CREATE TABLE role
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE employee
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    role_id    BIGINT REFERENCES role (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists employee;
drop table if exists role;
-- +goose StatementEnd
