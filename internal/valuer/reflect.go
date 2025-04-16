package valuer

import (
	"database/sql"
	"reflect"
	"web/orm/internal/errs"
	"web/orm/model"
)

var _ Creator = NewReflectValue

type reflectValue struct {
	model *model.Model
	// 对应T的指针
	val reflect.Value
}

func NewReflectValue(model *model.Model, val any) Value {
	return reflectValue{model: model, val: reflect.ValueOf(val).Elem()}
}

func (r reflectValue) Field(name string) (any, error) {
	return r.val.FieldByName(name).Interface(), nil
}

func (r reflectValue) SetColumn(rows *sql.Rows) error {

	// 如何构造 *T 并返回结果集

	// cs: 取出的列名
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	// 因为我们不知道用户会以什么样的顺序进行查询
	// 所以，我们构造一个any切片来存放对应的列名的类型的顺序
	vals := make([]any, 0, len(cs))
	valElems := make([]reflect.Value, 0, len(cs))

	// 遍历列名
	for _, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}

		// vals内存放着正确顺序的字段的零值
		val := reflect.New(fd.Type)
		vals = append(vals, val.Interface())
		valElems = append(valElems, val.Elem())
	}
	// 将数据库返回的当前行数据读取到传入的参数中
	//- 参数必须是指针类型，以便 Scan 可以修改它们的值
	//- 参数顺序必须与 SELECT 语句中的列顺序一致
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}

	tpValue := reflect.ValueOf(r.val)
	for k, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return nil
		}
		// 类似一个赋值操作
		tpValue.Elem().FieldByName(fd.GoName).
			Set(valElems[k])
	}
	return nil
}
