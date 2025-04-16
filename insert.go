package orm

import (
	"reflect"
	"strings"
	"web/orm/internal/errs"
	"web/orm/model"
)

type OnDuplicateKeyBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

type OnDuplicateKey struct {
	assigns         []Assignable
	conflictColumns []string
}

// ConflictColumns 这是一个中间方法，冲突列名
func (o *OnDuplicateKeyBuilder[T]) ConflictColumns(cols ...string) *OnDuplicateKeyBuilder[T] {
	o.conflictColumns = cols
	return o
}

// Update
// 大概用起来是这样：
// db.Insert(&user).OnDuplicateKey().Update(
//
//	Assign("age", 18),        // 直接赋值
//	C("name"),                // 使用 VALUES(name)
//
// )
func (o *OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicateKey = &OnDuplicateKey{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}

// Assignable 是一个用于处理赋值操作的接口，
// 主要用在 Upsert（INSERT ... ON DUPLICATE KEY UPDATE）场景中
type Assignable interface {
	assign()
}

type Inserter[T any] struct {
	// 定义成切片，是为了方便插入同一个结构体的多行列
	values []*T
	// 维持住DB是为了通过DB拿到一些信息
	db *DB
	//
	columns []string
	builder

	onDuplicateKey *OnDuplicateKey
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
			sb:      strings.Builder{},
		},
	}
}

func (i *Inserter[T]) OnDuplicateKey(assigns ...Assignable) *OnDuplicateKeyBuilder[T] {
	return &OnDuplicateKeyBuilder[T]{
		i: i,
	}
}

// Values 指定传入的参数并记录下来
func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

// Columns 指定要插入的列
func (i *Inserter[T]) Columns(col ...string) *Inserter[T] {
	i.columns = col
	return i
}

func (i *Inserter[T]) Build() (*Query, error) {
	n := len(i.values)
	if n == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	i.sb.WriteString("INSERT INTO ")

	// 拿到元数据
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	i.model = m

	// INSERT INTO `test_model`
	i.quote(m.TableName)

	// 用一个变量来代替m.Fields进行操作，防止m.Fields被污染
	// 而且操作更方便，减少了if else 的判断
	fields := m.Fields
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, fd := range i.columns {
			fdMeta, ok := m.FieldMap[fd]
			// 传入了乱七八糟的列
			if !ok {
				return nil, errs.NewErrUnknownField(fd)
			}
			fields = append(fields, fdMeta)
		}
	}

	// 显式指定列的顺序,不然我们不知道数据库中状认的顺序
	i.sb.WriteByte('(')
	for j, v := range fields {
		if j > 0 {
			i.sb.WriteByte(',')
		}
		//i.sb.WriteByte('`')
		//i.sb.WriteString(v.ColName)
		//i.sb.WriteByte('`')
		i.quote(v.ColName)
	}
	i.sb.WriteByte(')')

	i.sb.WriteString(" VALUES ")
	args := make([]any, 0, n*len(m.Fields))

	for j, val := range i.values {
		if j > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		// TODO 支持多列插入 大概要把下面提取成一个函数，然后遍历i.values，然后把sb内置成i的字段
		for idx, field := range fields {
			if idx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			// 在拥有字段的标识的时候，优先考虑直接用反射将对应的字段的值获取
			arg := reflect.ValueOf(val).Elem().FieldByName(field.GoName).Interface()
			args = append(args, arg)
		}
		i.sb.WriteString(")")
	}

	if i.onDuplicateKey != nil {
		err = i.dialect.buildOnDuplicateKey(&i.builder, i.onDuplicateKey)
		if err != nil {
			return nil, err
		}
		args = append(args, i.args...)
	}

	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
		Args: args,
	}, nil
}
