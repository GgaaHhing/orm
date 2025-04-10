package orm

import (
	"database/sql"
	"web/orm/internal/valuer"
	"web/orm/model"
)

type DBOption func(*DB)

// DB DB是sql.DB的一个装饰器
type DB struct {
	r       model.Registry
	db      *sql.DB
	creator valuer.Creator
}

func Open(driverName, dataSourceName string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		r:       model.NewRegistry(),
		db:      db,
		creator: valuer.NewUnsafeValue,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func DBUseReflect() DBOption {
	return func(r *DB) {
		r.creator = valuer.NewReflectValue
	}
}

func MistOpen(driverName, dataSourceName string, opts ...DBOption) *DB {
	res, err := Open(driverName, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return res
}
