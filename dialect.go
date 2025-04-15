package orm

import (
	"web/orm/internal/errs"
)

var (
	DialectMySOL  Dialect = mysqlDialect{}
	DialectSQLite Dialect = SQLiteDialect{}
)

// Dialect 方言抽象，用来迎合不同SQL的标准
type Dialect interface {
	// quoter 是为了解决不同SQL的引号问题
	quoter() byte
	// 构造OnDuplicateKey
	// 这里用 *builder是因为builder里面有strings.Builder
	buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error
}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	//TODO implement me
	panic("implement me")
}

func (s standardSQL) buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	//TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
}

func (s mysqlDialect) quoter() byte {
	return '`'
}

func (s mysqlDialect) buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=?")
			b.addArgs(a.val)

		// Column 是为了使用 col = VALUES(col)
		// 如果主键或唯一健冲突，则会将col的值改为你指定的col的值
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.sb.WriteByte('`')
			b.quote(fd.ColName)
			b.sb.WriteByte('`')
			b.sb.WriteByte(')')
		default:
			return errs.NewErrUnsupportedAssignable(assign)
		}
	}
	return nil
}

type SQLiteDialect struct {
	standardSQL
}
