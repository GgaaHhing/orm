package orm

import (
	"strings"
	"web/orm/internal/errs"
	"web/orm/model"
)

type Deleter[T any] struct {
	table string
	// 在where下面有各种条件
	where []Predicate
	model *model.Model
	sb    strings.Builder
	args  []any
	db    *DB
}

func NewDeleter[T any](db *DB) *Deleter[T] {
	return &Deleter[T]{
		db: db,
		sb: strings.Builder{},
	}
}

func (d *Deleter[T]) Build() (*Query, error) {
	// 先构造最基础的东西
	d.sb.WriteString("DELETE FROM ")
	// 解析model，获取表名
	var err error
	d.model, err = d.db.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	// 把表名加到里面
	if d.table != "" {
		d.sb.WriteByte('`')
		d.sb.WriteString(d.table)
		d.sb.WriteByte('`')
	} else {
		d.sb.WriteByte('`')
		d.sb.WriteString(d.model.TableName)
		d.sb.WriteByte('`')
	}

	// 串联，构造Where语句
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		// 然后进行串联
		p := d.where[0]
		for _, w := range d.where[1:] {
			p = p.And(w)
		}

		// 串联完成之后，进行构造环节
		if err = d.buildExpression(p); err != nil {
			return nil, err
		}

		d.sb.WriteByte(';')
	} else {
		return &Query{
			SQL: d.sb.String(),
		}, errs.ErrDeleteALL
	}

	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

// buildExpression
// 我们传入的是Predicate 或者 Column，它们都实现了Expression接口，顶级抽象
func (d *Deleter[T]) buildExpression(expr Expression) error {
	switch exp := expr.(type) {
	case nil:
	// C("id").Eq(12).And(C("name").Eq("Tom"))
	// `id` = 12 AND `name` = "Tom"
	case Predicate:
		// 构造完左边的Predicate
		_, ok := exp.left.(Predicate)
		if ok {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			d.sb.WriteByte(')')
		}

		// 构造中间的表达式或者是 OP
		d.sb.WriteString(" " + exp.op.String() + " ")

		// 最后构造右边的Predicate
		_, ok = exp.right.(Predicate)
		if ok {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			d.sb.WriteByte(')')
		}

	// 如果是一个列名：就构造成 `age` =
	case Column:
		d.sb.WriteByte('`')
		fd, ok := d.model.FieldMap[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		d.sb.WriteString(fd.ColName)
		d.sb.WriteByte('`')

	// 如果解析到最后，发现是一个参数，我们就要存储起来
	case value:
		d.args = append(d.args, exp.val)
		d.sb.WriteByte('?')

	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

// Where 让用户传表达式进来，然后我们自己构造
func (d *Deleter[T]) Where(ps ...Predicate) *Deleter[T] {
	d.where = ps
	return d
}

func (d *Deleter[T]) From(table string) *Deleter[T] {
	d.table = table
	return d
}
