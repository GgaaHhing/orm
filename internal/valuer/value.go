package valuer

import (
	"database/sql"
	"web/orm/model"
)

type Value interface {
	// Field 读取字段
	Field(name string) (any, error)
	SetColumn(rows *sql.Rows) error
}

// Creator 构造Value抽象
type Creator func(model *model.Model, entity any) Value
