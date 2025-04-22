package orm

import (
	"context"
	"errors"
	"reflect"
	"strings"
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
	table TableReference
	// 在where下面有各种条件
	where []Predicate

	// 分列查询
	columns []Selectable
	groupBy []Column    // 添加 groupBy 字段
	having  []Predicate // 添加 having 字段

	core
	sess Session
}

func NewSelector[T any](sess Session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		sess: sess,
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
		core: c,
	}
}

// Build 解析字段，构造对应的查询语句
func (s *Selector[T]) Build() (*Query, error) {
	var err error
	if s.model == nil {
		s.model, err = s.r.Get(new(T))
		if err != nil {
			return nil, err
		}
	}
	s.sb.WriteString("SELECT ")
	err = s.buildColumns()
	if err != nil {
		return nil, err
	}

	s.sb.WriteString(" FROM ")

	if err = s.buildTable(s.table); err != nil {
		return nil, err
	}

	//if s.table != "" {
	//	// 防止用户传一些嵌套表名之类的
	//	// 干脆如果用户有特殊需求，就自己传表名和反引号
	//	//s.sb.WriteByte('`')
	//	s.sb.WriteString(s.table)
	//	//s.sb.WriteByte('`')
	//} else {
	//	s.sb.WriteByte('`')
	//	s.sb.WriteString(s.model.TableName)
	//	s.sb.WriteByte('`')
	//}

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

	// WHERE 子句之后 Group By
	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, c := range s.groupBy {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			if err = s.buildColumn(c); err != nil {
				return nil, err
			}
		}
	}

	// GROUP BY 之后添加 HAVING
	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		p := s.having[0]
		for i := 1; i < len(s.having); i++ {
			p = p.And(s.having[i])
		}
		// HAVING COUNT(`age`) > ?
		// Aggregate opGt value
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

func (s *Selector[T]) buildTable(table TableReference) error {
	switch t := table.(type) {
	case nil:
		// 没有调用From，使用默认的逻辑
		s.quote(s.model.TableName)
	case Table:
		// 解析元数据
		m, err := s.r.Get(t.entity)
		if err != nil {
			return err
		}
		s.quote(m.TableName)
		if t.alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(t.alias)
		}
	case Join:
		s.sb.WriteByte('(')
		// 构造右边
		err := s.buildTable(t.left)
		if err != nil {
			return err
		}
		s.sb.WriteString(t.typ)
		err = s.buildTable(t.right)
		if err != nil {
			return err
		}
		s.sb.WriteByte(')')

		if len(t.using) > 0 {
			s.sb.WriteString(" USING (")
			for i, col := range t.using {
				if i > 0 {
					s.sb.WriteByte(',')
				}
				err = s.buildColumn(Column{name: col})
				if err != nil {
					return err
				}
			}
			s.sb.WriteString(")")
		}

		if len(t.on) > 0 {
			s.sb.WriteString(" ON ")
			p := t.on[0]
			for i := 1; i < len(t.on); i++ {
				p = p.And(t.on[i])
			}
			// HAVING COUNT(`age`) > ?
			// Aggregate opGt value
			if err = s.buildExpression(p); err != nil {
				return err
			}
		}

	default:
		return errs.NewErrUnsupportedTable(table)
	}
	return nil
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
		s.addArgs(exp.val)
		s.sb.WriteString("?")

	case RawExpr:
		if len(exp.args) > 0 {
			s.addArgs(exp.args...)
		}
		// 检查是否是聚合函数表达式
		if isAggregateExpr(exp.raw) {
			s.sb.WriteString(exp.raw)
		} else {
			// 用户自定义的 RawExpr，添加括号保证优先级
			s.sb.WriteByte('(')
			s.sb.WriteString(exp.raw)
			s.sb.WriteByte(')')
		}

	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

// 判断是否是聚合函数表达式
func isAggregateExpr(expr string) bool {
	aggregateFuncs := []string{"COUNT", "SUM", "AVG", "MAX", "MIN"}
	for _, fn := range aggregateFuncs {
		if strings.HasPrefix(expr, fn) {
			return true
		}
	}
	return false
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
				s.addArgs(c.args...)
			}

		default:
			return errors.New("")
		}
	}
	return nil
}

func (s *Selector[T]) From(table TableReference) *Selector[T] {
	s.table = table
	return s
}

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
	var err error
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	res := get[T](ctx, s.sess, s.core, &QueryContext{
		Type:    "SELECT",
		Builder: s,
		Model:   s.model,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

// Get 将对应的查询语句发给数据库并接收返回的查询结果
//func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
//	var err error
//	s.model, err = s.r.Get(new(T))
//	if err != nil {
//		return nil, err
//	}
//	root := s.getHandler
//	for i := len(s.mdls) - 1; i >= 0; i-- {
//		root = s.mdls[i](root)
//	}
//	res := root(ctx, &QueryContext{
//		Type:    "SELECT",
//		Builder: s,
//		Model:   s.model,
//	})
//	if res.Result != nil {
//		return res.Result.(*T), res.Err
//	}
//	return nil, res.Err
//}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}

	rows, err := s.sess.queryContext(ctx, q.SQL, q.Args...)
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

// GroupBy 设置 GROUP BY 子句
func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupBy = cols
	return s
}

// Having 设置 HAVING 子句
func (s *Selector[T]) Having(ps ...Predicate) *Selector[T] {
	s.having = ps
	return s
}
