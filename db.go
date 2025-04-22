package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"log"
	"time"
	"web/orm/internal/errs"
	"web/orm/internal/valuer"
	"web/orm/model"
)

type DBOption func(*DB)

// DB DB是sql.DB的一个装饰器
type DB struct {
	core
	//r       model.Registry
	db *sql.DB
	//creator valuer.Creator
	//dialect Dialect
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
		core: core{
			r:       model.NewRegistry(),
			creator: valuer.NewUnsafeValue,
			dialect: DialectMySOL,
		},
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func (db *DB) getCore() core {
	return db.core
}

// DoTx 事务闭包
func (db *DB) DoTx(ctx context.Context, fn func(ctx context.Context, tx *Tx) error,
	opts *sql.TxOptions) (err error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return
	}
	panicked := true
	defer func() {
		if panicked || err != nil {
			e := tx.Rollback()
			err = errs.NewErrFailedToRollbackTx(err, e, panicked)
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(ctx, tx)
	panicked = false
	return
}

// BeginTx 开启事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

func DBWithDialect(dialect Dialect) DBOption {
	return func(r *DB) {
		r.dialect = dialect
	}
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

func DBWithMiddleware(mdls ...Middleware) DBOption {
	return func(r *DB) {
		r.mdls = mdls
	}
}

// Wait 集成测试不会等待mysql的启动，所以需要加入wait确保mysql启动成功
func (db *DB) Wait() error {
	err := db.db.Ping()
	for errors.Is(err, driver.ErrBadConn) {
		log.Printf("等待数据库启动。。")
		err = db.db.Ping()
		time.Sleep(time.Second)
	}
	return err
}
