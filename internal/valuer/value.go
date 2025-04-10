package valuer

import (
	"database/sql"
	"web/orm/model"
)

type Value interface {
	SetColumn(rows *sql.Rows) error
}

// Creator 构造抽象
type Creator func(model *model.Model, entity any) Value
