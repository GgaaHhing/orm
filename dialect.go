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
	buildOnDuplicateKey(b *builder, odk *Upsert) error
}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	//TODO implement me
	panic("implement me")
}

func (s standardSQL) buildOnDuplicateKey(b *builder, odk *Upsert) error {
	//TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
}

func (m mysqlDialect) quoter() byte {
	return '`'
}

func (m mysqlDialect) buildOnDuplicateKey(b *builder, odk *Upsert) error {
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

		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.quote(fd.ColName)
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

// buildOnDuplicateKey SQLite的语法大概是：
// INSERT INTO table_name (column1, column2)
// VALUES (value1, value2)
// ON CONFLICT (conflict_column) DO UPDATE SET
//
//	column1 = excluded.column1,
//	column2 = value2;
//
// 使用 excluded 关键字引用插入的新值（相当于 MySQL 的 VALUES ）
func (s SQLiteDialect) buildOnDuplicateKey(b *builder, odk *Upsert) error {
	b.sb.WriteString("ON CONFLICT (")
	for i, col := range odk.conflictColumns {
		if i > 0 {
			b.sb.WriteByte(',')
		}
		err := b.buildColumn(col)
		if err != nil {
			return err
		}
	}
	b.sb.WriteString(") DO UPDATE SET ")
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
			b.sb.WriteString("=excluded.")
			b.quote(fd.ColName)
		default:
			return errs.NewErrUnsupportedAssignable(assign)
		}
	}
	return nil
}
