package orm

// op本身应该是枚举，但定义成string的衍生类型更方便
const (
	opEq  op = "="
	opNeq op = "!="
	opLt  op = "<"
	opGt  op = ">"
	opNot op = "NOT"
	opAnd op = "AND"
	opOr  op = "OR"
)

// Predicate 查询条件结构体
type Predicate struct {
	left  Expression
	op    op
	right Expression
}

type op string

func (o op) String() string {
	return string(o)
}

func (left Predicate) expr() {}

// Not Not后面接的是条件，所以需要传入Predicate
// 大概用法：Not(C("id).Eq(12))
func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: p,
	}
}

// And
// 大概用法：C("id").Eq(12).And(C("name").Eq("Tom").And(xxx))
func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opAnd,
		right: right,
	}
}

func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOr,
		right: right,
	}
}

type value struct {
	val any
}

func (value) expr() {}
