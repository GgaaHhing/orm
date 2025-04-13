package orm

// Expression 标记接口代表表达式
// 定义了一个共同的抽象，使得不同类型的表达式可以统一处理
type Expression interface {
	expr()
}

// RawExpr 表示原生表达式
type RawExpr struct {
	raw  string
	args []any
}

func Raw(raw string, args ...any) RawExpr {
	return RawExpr{
		raw:  raw,
		args: args,
	}
}

// expr 实现这个抽象是为了支持Where语句
func (r RawExpr) expr() {}

// selectable 实现这个抽象是为了支持Select 的 XXX 查询
func (r RawExpr) selectable() {}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}
