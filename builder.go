package orm

import (
	"strings"
	"web/orm/internal/errs"
)

// builder 用于构造不同抽象之间的公共部分的builder
// 例如 `col`这样的构造，model的流通，等等......
type builder struct {
	sb   strings.Builder
	args []any
	core
	quoter byte
}

// quote 构造列名 `col`
func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

// buildColumn 如果是列名，我们就把它构造成 `age` 这样的形式
func (b *builder) buildColumn(col Column) error {
	switch table := col.table.(type) {
	case nil:
		fd, ok := b.model.FieldMap[col.name]
		if !ok {
			return errs.NewErrUnknownField(col.alias)
		}
		//b.quote(fd.ColName)
		b.sb.WriteByte('`')
		b.sb.WriteString(fd.ColName)
		b.sb.WriteByte('`')
		if col.alias != "" {
			b.sb.WriteString(" AS `")
			b.sb.WriteString(col.alias)
			b.sb.WriteByte('`')
		}
	case Table:
		m, err := b.r.Get(table.entity)
		if err != nil {
			return err
		}
		fd, ok := m.FieldMap[col.name]
		if !ok {
			return errs.NewErrUnknownField(col.alias)
		}
		if table.alias != "" {
			b.quote(table.alias)
			b.sb.WriteByte('.')
		}
		b.quote(fd.ColName)
		if col.alias != "" {
			b.sb.WriteString(" AS ")
			b.quote(col.alias)
		}
	default:
		return errs.NewErrUnsupportedTable(table)
	}
	return nil
}

//func (b *builder) buildColumn(name string) error {
//	fd, ok := b.model.FieldMap[name]
//	if !ok {
//		return errs.NewErrUnknownField(name)
//	}
//	b.quote(fd.ColName)
//	return nil
//}

func (b *builder) addArgs(args ...any) {
	if len(args) == 0 {
		return
	}
	if b.args == nil {
		b.args = make([]any, 0, 4)
	}
	b.args = append(b.args, args...)
}
