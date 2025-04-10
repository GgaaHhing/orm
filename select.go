package orm

import (
	"context"
	"reflect"
	"strings"
	"web/orm/internal/errs"
	"web/orm/model"
)

type Selector[T any] struct {
	table string
	// 在where下面有各种条件
	where []Predicate
	model *model.Model
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
		s.sb.WriteString(s.model.TableName)
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
		fd, ok := s.model.FieldMap[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.ColName)
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
	tp := new(T)
	val := s.db.creator(s.model, tp)
	err = val.SetColumn(rows)
	return tp, err
}

//func (s *Selector[T]) GetV1(ctx context.Context) (*T, error) {
//	q, err := s.Build()
//	if err != nil {
//		return nil, err
//	}
//
//	db := s.db.db
//	rows, err := db.QueryContext(ctx, q.SQL, q.Args)
//	if err != nil {
//		return nil, err
//	}
//
//	if !rows.Next() {
//		return nil, ErrNoRows
//	}
//	tp := new(T)
//	//var creator valuer.Creator
//	//val := creator(tp)
//	//val.SetColumn(rows)
//
//	return tp, err
//	// cs: 取出的列名
//	//cs, err := rows.Columns()
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//var vals []any
//	//address := reflect.ValueOf(tp).UnsafePointer()
//	//for _, c := range cs {
//	//	fd, ok := s.model.ColumnMap[c]
//	//	if !ok {
//	//		return nil, errs.NewErrUnknownColumn(c)
//	//	}
//	//	fdAddr := unsafe.Pointer(uintptr(address) + fd.Offset)
//	//	// Scan需要指针类型，所以这里不需要加Elem
//	//	val := reflect.NewAt(fd.typ, fdAddr) // .Elem()
//	//	vals = append(vals, val.Interface())
//	//}
//	//
//	//err = rows.Scan(vals...)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//return tp, nil
//}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}

	db := s.db.db

	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 用来装数据
	var result []*T
	for rows.Next() {
		// 我觉得用户的顺序可能混乱，所以我需要知道用户的查询顺序
		cs, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		vals := make([]any, 0, len(cs))
		valsElem := make([]reflect.Value, 0, len(cs))

		for _, c := range cs {
			// 获取的fd就是我们的字段名
			fd, ok := s.model.ColumnMap[c]
			if !ok {
				return nil, errs.NewErrUnknownField(c)
			}
			// 返回一个初始化为具体值的新值
			val := reflect.New(fd.Type)
			vals = append(vals, val.Interface())
			// 当你对通过 reflect.New() 创建的值调用 Elem() 时，它返回指针指向的值
			valsElem = append(valsElem, val.Elem())
		}
		// rows.Scan(vals...) 执行后，valsElem中的这些值会被填充为数据库返回的实际数据。
		err = rows.Scan(vals...)
		if err != nil {
			return nil, err
		}
		// id name age
		// name id age
		// cs = name, id, age
		// vals = [name, id, age]

		// 获取了顺序之后，我们就需要将获得的数据写入到指定的结构体里
		tp := new(T)
		// 获取用户传入的结构体
		tpValue := reflect.ValueOf(tp)
		for k, c := range cs {
			fd, ok := s.model.ColumnMap[c]
			if !ok {
				return nil, errs.NewErrUnknownField(c)
			}

			tpValue.Elem().FieldByName(fd.GoName).Set(valsElem[k])
		}
		result = append(result, tp)
	}
	return result, nil
}
