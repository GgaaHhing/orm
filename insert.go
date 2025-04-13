package orm

import "strings"

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
	cnt := 0
	for _, v := range m.FieldMap {
		if cnt > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('`')
		sb.WriteString(v.ColName)
		sb.WriteByte('`')
		cnt++
	}
	sb.WriteByte(')')
	return &Query{
		SQL: sb.String(),
	}, nil
}
