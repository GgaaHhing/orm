package orm

type DBOption func(*DB)

type DB struct {
	r *registry
}

func NewDB(opts ...DBOption) (*DB, error) {
	res := &DB{
		r: NewRegistry(),
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func MistNewDB(opts ...DBOption) *DB {
	res, err := NewDB(opts...)
	if err != nil {
		panic(err)
	}
	return res
}
