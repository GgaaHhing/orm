package orm

const (
	opEq  op = "="
	opNot op = "NOT"
)

type Predicate struct {
	c   Column
	op  op
	arg any
}

type op string

type Column struct {
	name string
}

func C(name string) Column {
	return Column{name: name}
}

func (c Column) Eq(arg any) Predicate {
	return Predicate{
		c:   c,
		op:  opEq,
		arg: arg,
	}
}

func Not(p Predicate) Predicate {
	return Predicate{
		op:  opNot,
		arg: p,
	}
}
