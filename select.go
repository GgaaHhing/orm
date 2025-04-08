package orm

import (
	"context"
	"reflect"
	"strings"
	"web/orm/internal/errs"
)

type Selector[T any] struct {
	table string
	// 在where下面有各种条件
	where []Predicate
	model *Model
	sb    strings.Builder
	args  []any

	db *DB
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
		sb: strings.Builder{},
	}
}

// Build 解析字段，构造对应的查询语句
func (s *Selector[T]) Build() (*Query, error) {
	s.sb.WriteString("SELECT * FROM ")

	// 解析model
	var err error
	s.model, err = s.db.r.Get(new(T))
	if err != nil {
		return nil, err
	}

	if s.table != "" {
		// 防止用户传一些嵌套表名之类的
		// 干脆如果用户有特殊需求，就自己传表名和反引号
		//s.sb.WriteByte('`')
		s.sb.WriteString(s.table)
		//s.sb.WriteByte('`')
	} else {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.tableName)
		s.sb.WriteByte('`')
	}

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		// 拼接
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		if err = s.buildExpression(p); err != nil {
			return nil, err
		}
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildExpression(expr Expression) error {
	switch exp := expr.(type) {
	case nil:
	case Predicate:
		_, ok := exp.left.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

		s.sb.WriteString(" " + exp.op.String() + " ")

		// WHERE (`Age` = ?) AND (`name` = ?)
		_, ok = exp.right.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

	case Column:
		fd, ok := s.model.fieldMap[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.colName)
		s.sb.WriteByte('`')

	case value:
		s.addArg(exp.val)
		s.sb.WriteString("?")

	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

func (s *Selector[T]) addArg(val any) *Selector[T] {
	if s.args == nil {
		s.args = make([]any, 0, 4)
	}
	s.args = append(s.args, val)
	return s
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
}

func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

// Get 将对应的查询语句发给数据库并接收返回的查询结果
func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	// 构造查询
	q, err := s.Build()
	if err != nil {
		return nil, err
	}

	db := s.db.db
	rows, err := db.QueryContext(ctx, q.SQL, q.Args)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, ErrNoRows
	}
	// 如何构造 *T 并返回结果集

	// cs: 取出的列名
	cs, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 因为我们不知道用户会以什么样的顺序进行查询
	// 所以，我们构造一个any切片来存放对应的列名的类型的顺序
	vals := make([]any, 0, len(cs))
	valElems := make([]reflect.Value, 0, len(cs))

	// 遍历列名
	for _, c := range cs {
		fd, ok := s.model.columnMap[c]
		if !ok {
			return nil, errs.NewErrUnknownColumn(c)
		}

		// vals内存放着正确顺序的字段的零值
		val := reflect.New(fd.typ)
		vals = append(vals, val.Interface())
		valElems = append(valElems, val.Elem())
	}
	// 将数据库返回的当前行数据读取到传入的参数中
	//- 参数必须是指针类型，以便 Scan 可以修改它们的值
	//- 参数顺序必须与 SELECT 语句中的列顺序一致
	err = rows.Scan(vals...)
	if err != nil {
		return nil, err
	}

	// new返回的是指针
	tp := new(T)
	tpValue := reflect.ValueOf(tp)
	for k, c := range cs {
		fd, ok := s.model.columnMap[c]
		if !ok {
			return nil, errs.NewErrUnknownColumn(c)
		}
		// 类似一个赋值操作
		tpValue.Elem().FieldByName(fd.goName).
			Set(valElems[k])
	}
	return tp, nil
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}

	db := s.db.db

	rows, err := db.QueryContext(ctx, q.SQL, q.Args)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	for rows.Next() {

	}
	panic("implement me")
}
