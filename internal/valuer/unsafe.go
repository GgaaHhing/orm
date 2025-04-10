package valuer

import (
	"database/sql"
	"reflect"
	"unsafe"
	"web/orm/internal/errs"
	"web/orm/model"
)

var _ Creator = NewUnsafeValue

type UnsafeValue struct {
	model *model.Model
	val   any
}

func NewUnsafeValue(model *model.Model, val any) Value {
	return UnsafeValue{model: model, val: val}
}

func (u UnsafeValue) SetColumn(rows *sql.Rows) error {
	//cs: 取出的列名
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	var vals []any
	address := reflect.ValueOf(u.val).UnsafePointer()
	for _, c := range cs {
		fd, ok := u.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		fdAddr := unsafe.Pointer(uintptr(address) + fd.Offset)
		// Scan需要指针类型，所以这里不需要加Elem
		val := reflect.NewAt(fd.Type, fdAddr) // .Elem()
		vals = append(vals, val.Interface())
	}

	err = rows.Scan(vals...)
	return err
}
