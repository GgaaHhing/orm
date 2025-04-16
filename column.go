package orm

type Column struct {
	name  string
	alias string
}

func C(name string) Column {
	return Column{name: name}
}

// Column ：用于 VALUES 语法，如 col = VALUES(col)
func (c Column) assign() {}

func (c Column) As(alias string) Column {
	return Column{name: c.name, alias: alias}
}

// Eq =
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left: c,
		op:   opEq,
		// 这里暂时无法解决对于any和Expression的关联
		// 所以，直接新建一个struct来关联Expression
		//right: arg,
		right: valueOf(arg),
	}
}

func valueOf(arg any) Expression {
	switch v := arg.(type) {
	case Expression:
		return v
	default:
		return value{val: v}
	}
}

func (c Column) selectable() {}

func (Column) expr() {}
