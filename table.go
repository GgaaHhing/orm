package orm

type TableReference interface {
	table()
}

// 定义不同的表，然后改造From，让他接收这个顶级抽象，然后区分类别，构造不同的语句

// Table 普通表
// 大概用法: Table-JoinBuilder-Join.xxx()
type Table struct {
	entity any
	alias  string
}

func TableOf(entity any) Table {
	return Table{entity: entity}
}

func (t Table) As(alias string) Table {
	return Table{
		entity: t.entity,
		alias:  alias,
	}
}

// C 用于指定某表的某个列
func (t Table) C(name string) Column {
	return Column{
		name:  name,
		table: t,
	}
}

func (t Table) table() {
	//TODO implement me
	panic("implement me")
}

func (t Table) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   " JOIN ",
	}
}

func (t Table) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   " LEFT JOIN ",
	}
}

func (t Table) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   " RIGHT JOIN ",
	}
}

// Join 连接
// Join的几种写法：
/*
	INNER JOIN（内连接）
	LEFT JOIN（左连接）
	RIGHT JOIN（右连接）
	JOIN ..... ON 表达式
	JOIN ..... USING(列名)
*/
type Join struct {
	left  TableReference
	right TableReference
	typ   string
	on    []Predicate
	using []string
}

func (j Join) table() {
	//TODO implement me
	panic("implement me")
}

func (j Join) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   " JOIN ",
	}
}

func (j Join) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   " LEFT JOIN ",
	}
}

func (j Join) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   " RIGHT JOIN ",
	}
}

// JoinBuilder 作为一个粘合剂，连接左边的查询语句和右边的查询语句
type JoinBuilder struct {
	left  TableReference
	right TableReference
	// typ 用来区分是rightJoin还是leftJoin
	typ string
}

// On table.Join(xxx).On(C("Id").Eq(12))
func (j *JoinBuilder) On(ps ...Predicate) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		on:    ps,
	}
}

func (j *JoinBuilder) Using(cols ...string) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		using: cols,
	}
}
