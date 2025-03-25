package orm

import (
	"context"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	table string
	where []Predicate
}

func (s *Selector[T]) Build() (*Query, error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	if s.table != "" {
		// 防止用户传一些嵌套表名之类的
		// 干脆如果用户有特殊需求，就自己传表名和反引号
		//sb.WriteByte('`')
		sb.WriteString(s.table)
		//sb.WriteByte('`')
	} else {
		var t T
		typ := reflect.TypeOf(t)
		sb.WriteByte('`')
		sb.WriteString(typ.Name())
		sb.WriteByte('`')
	}
	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: nil,
	}, nil
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
