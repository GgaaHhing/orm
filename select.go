package orm

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	table string
	// 在where下面有各种条件
	where []Predicate
	sb    strings.Builder
	args  []any
}

func (s *Selector[T]) Build() (*Query, error) {
	s.sb.WriteString("SELECT * FROM ")
	if s.table != "" {
		// 防止用户传一些嵌套表名之类的
		// 干脆如果用户有特殊需求，就自己传表名和反引号
		//s.sb.WriteByte('`')
		s.sb.WriteString(s.table)
		//s.sb.WriteByte('`')
	} else {
		var t T
		typ := reflect.TypeOf(t)
		s.sb.WriteByte('`')
		s.sb.WriteString(typ.Name())
		s.sb.WriteByte('`')
	}

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		// 拼接
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		if err := s.buildExpression(p); err != nil {
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
		s.sb.WriteByte('`')
		s.sb.WriteString(exp.name)
		s.sb.WriteByte('`')

	case value:
		s.addArg(exp.val)
		s.sb.WriteString("?")

	default:
		return fmt.Errorf("orm: 不支持的表达式类型 %v", expr)
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
