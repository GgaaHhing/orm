package orm

import (
	"reflect"
	"strings"
	"web/orm/internal/errs"
)

type Inserter[T any] struct {
	// 定义成切片，是为了方便插入同一个结构体的多行列
	values []*T
	// 维持住DB是为了通过DB拿到一些信息
	db *DB
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
	}
}

// Values 指定传入的参数并记录下来
func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) Build() (*Query, error) {
	n := len(i.values)
	if n == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")

	// 拿到元数据
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}

	// INSERT INTO `test_model`
	sb.WriteByte('`')
	sb.WriteString(m.TableName)
	sb.WriteByte('`')

	// 显式指定列的顺序,不然我们不知道数据库中状认的顺序
	sb.WriteByte('(')
	for j, v := range m.Fields {
		if j > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('`')
		sb.WriteString(v.ColName)
		sb.WriteByte('`')
	}
	sb.WriteByte(')')

	sb.WriteString(" VALUES ")
	args := make([]any, 0, n*len(m.Fields))

	for j, val := range i.values {
		if j > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('(')
		// TODO 支持多列插入 大概要把下面提取成一个函数，然后遍历i.values，然后把sb内置成i的字段
		for idx, field := range m.Fields {
			if idx > 0 {
				sb.WriteByte(',')
			}
			sb.WriteByte('?')
			// 在拥有字段的标识的时候，优先考虑直接用反射将对应的字段的值获取
			arg := reflect.ValueOf(val).Elem().FieldByName(field.GoName).Interface()
			args = append(args, arg)
		}
		sb.WriteString(")")
	}
	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: args,
	}, nil
}
