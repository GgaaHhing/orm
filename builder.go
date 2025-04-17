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

func (b *builder) buildColumn(name string) error {
	fd, ok := b.model.FieldMap[name]
	if !ok {
		return errs.NewErrUnknownField(name)
	}
	b.quote(fd.ColName)
	return nil
}

func (b *builder) addArgs(args ...any) {
	if len(args) == 0 {
		return
	}
	if b.args == nil {
		b.args = make([]any, 0, 4)
	}
	b.args = append(b.args, args...)
}
