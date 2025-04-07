package orm

import "database/sql"

type DBOption func(*DB)

// DB DB是sql.DB的一个装饰器
type DB struct {
	r  *registry
	db *sql.DB
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
		r:  NewRegistry(),
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func MistOpen(driverName, dataSourceName string, opts ...DBOption) *DB {
	res, err := Open(driverName, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return res
}
