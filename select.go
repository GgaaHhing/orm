package orm

import (
	"context"
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

func (s *Selector[T]) Build() (*Query, error) {
	s.sb.WriteString("SELECT * FROM ")

	// 解析model
	var err error
	s.model, err = s.db.r.get(new(T))
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
		fd, ok := s.model.fields[exp.name]
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

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}
