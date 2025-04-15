package orm

import (
	"context"
	"errors"
	"reflect"
	"web/orm/internal/errs"
)

// Selectable 是一个标记接口
// 它代表的是查找的列，或者聚合函数等
// SELECT xxX 部分
type Selectable interface {
	selectable()
}

type Selector[T any] struct {
	builder
	table string
	// 在where下面有各种条件
	where []Predicate

	// 分列查询
	columns []Selectable

	db *DB
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
	}
}

// Build 解析字段，构造对应的查询语句
func (s *Selector[T]) Build() (*Query, error) {
	var err error
	s.sb.WriteString("SELECT ")
	err = s.buildColumns()
	if err != nil {
		return nil, err
	}

	s.sb.WriteString(" FROM ")

	// 解析model
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
	// 如果是nil就执行就什么都不做
	case nil:
	// 如果是Predicate，说明是表达式，用递归不断筛选出合适的进行构造
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

		if exp.op != "" {
			s.sb.WriteString(" " + exp.op.String() + " ")
		}

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
		// TODO
		// 忽略别名
		exp.alias = ""
		return s.buildColumn(exp)

	// 如果是值，我们就把它添加进s中，然后用占位符表示
	// 防止SQL注入
	case value:
		s.addArg(exp.val)
		s.sb.WriteString("?")

	case RawExpr:
		if len(exp.args) > 0 {
			s.addArg(exp.args...)
		}
		s.sb.WriteByte('(')
		s.sb.WriteString(exp.raw)
		s.sb.WriteByte(')')

	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sb.WriteByte('*')
		return nil
	}
	for i, col := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch c := col.(type) {
		case Column:
			err := s.buildColumn(c)
			if err != nil {
				return err
			}

		case Aggregate:
			s.sb.WriteString(c.fn)
			s.sb.WriteByte('(')
			// TODO
			err := s.buildColumn(Column{name: c.arg})
			if err != nil {
				return err
			}
			s.sb.WriteByte(')')

			if c.alias != "" {
				s.sb.WriteString(" AS `")
				s.sb.WriteString(c.alias)
				s.sb.WriteByte('`')
			}

		case RawExpr:
			s.sb.WriteString(c.raw)
			if len(c.args) > 0 {
				s.addArg(c.args...)
			}

		default:
			return errors.New("")
		}
	}
	return nil
}

// buildColumn 如果是列名，我们就把它构造成 `age` 这样的形式
func (s *Selector[T]) buildColumn(col Column) error {
	fd, ok := s.model.FieldMap[col.name]
	if !ok {
		return errs.NewErrUnknownField(col.alias)
	}
	s.sb.WriteByte('`')
	s.sb.WriteString(fd.ColName)
	s.sb.WriteByte('`')
	if col.alias != "" {
		s.sb.WriteString(" AS `")
		s.sb.WriteString(col.alias)
		s.sb.WriteByte('`')
	}
	return nil
}

// addArg 为Selector添加参数
func (s *Selector[T]) addArg(val ...any) {
	if len(val) == 0 {
		return
	}
	if s.args == nil {
		s.args = make([]any, 0, 4)
	}
	s.args = append(s.args, val...)
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
}

//func (s *Selector[T]) Select(cols ...string) *Selector[T] {
//	s.columns = cols
//	return s
//}

func (s *Selector[T]) Selectable(cols ...Selectable) *Selector[T] {
	s.columns = cols
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
