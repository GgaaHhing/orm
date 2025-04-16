package orm

import "web/orm/model"

type Aggregate struct {
	fn    string
	arg   string
	alias string
}

func (a Aggregate) expr() {}

func (a Aggregate) selectable() {}

func (a Aggregate) As(alias string) Aggregate {
	return Aggregate{
		fn:    a.fn,
		arg:   a.arg,
		alias: alias,
	}
}

func (a Aggregate) Gt(arg any) Predicate {
	return Predicate{
		left: RawExpr{
			// 移除多余的括号，直接使用聚合函数表达式
			raw:  a.fn + "(`" + model.UnderscoreCase(a.arg) + "`)",
			args: nil,
		},
		op:    opGt,
		right: valueOf(arg),
	}
}

func Avg(col string) Aggregate {
	return Aggregate{
		fn:  "AVG",
		arg: col,
	}
}

func Sum(col string) Aggregate {
	return Aggregate{
		fn:  "SUM",
		arg: col,
	}
}

func Count(col string) Aggregate {
	return Aggregate{
		fn:  "COUNT",
		arg: col,
	}
}

func Max(col string) Aggregate {
	return Aggregate{
		fn:  "MAX",
		arg: col,
	}
}

func Min(col string) Aggregate {
	return Aggregate{
		fn:  "MIN",
		arg: col,
	}
}
