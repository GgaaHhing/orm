package orm

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"web/orm/internal/errs"
)

const (
	// 在标签里 column专门用于重命名列名
	tagKeyColumn = "column"
)

var (
	matchFirstCap            = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap              = regexp.MustCompile("([a-z0-9])([A-Z])")
	matchMultipleUnderscores = regexp.MustCompile("_+")
)

// Registry 我们希望可以提供一些拓展性给Model，让用户可以自定义
type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...ModelOption) (*Model, error)
}

type Model struct {
	tableName string
	// 字段
	fieldMap map[string]*Field
	// 列
	columnMap map[string]*Field
}

type ModelOption func(*Model) error

type Field struct {
	// Go中的名字
	goName string
	// 列名
	colName string

	typ reflect.Type
}

// registry 元数据注册中心
type registry struct {
	lock   sync.RWMutex
	models map[reflect.Type]*Model
}

func NewRegistry() *registry {
	return &registry{
		models: make(map[reflect.Type]*Model, 64),
	}
}

func ModelWithTableName(tableName string) ModelOption {
	return func(r *Model) error {
		r.tableName = tableName
		if tableName == "" {
			return errors.New("orm: table name is empty")
		}
		return nil
	}
}

func ModelWithColumnName(field, colName string) ModelOption {
	return func(r *Model) error {
		fd, ok := r.fieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		fd.colName = colName
		return nil
	}
}

func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	r.lock.RLock()
	m, ok := r.models[typ]
	r.lock.RUnlock()
	if ok {
		return m, nil
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	m, ok = r.models[typ]
	if ok {
		return m, nil
	}

	var err error
	// 如果不ok，说明是我没解析过的，我解析一下
	m, err = r.Register(val)
	if err != nil {
		return nil, err
	}
	r.models[typ] = m
	return m, nil
}

func (r *registry) Register(entity any, opts ...ModelOption) (*Model, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()
	numField := typ.NumField()
	fieldMap := make(map[string]*Field, numField)
	columnMap := make(map[string]*Field, numField)
	for i := 0; i < numField; i++ {
		f := typ.Field(i)
		// pair中包含了结构体中目前字段解析出来的tag
		pair, err := r.parseTag(f.Tag)
		if err != nil {
			return nil, err
		}
		colName := pair[tagKeyColumn]
		// 如果标签为空，我们就帮用户进行处理
		if colName == "" {
			colName = underscoreCase(f.Name)
		}
		fd := &Field{
			colName: colName,
			goName:  f.Name,
			typ:     f.Type,
		}
		fieldMap[f.Name] = fd
		// column就是用户自定义的字段名称
		columnMap[colName] = fd
	}

	var tableName string
	if tbl, ok := entity.(TableName); ok {
		tableName = tbl.TableName()
	}
	if tableName == "" {
		tableName = underscoreCase(typ.Name())
	}

	res := &Model{
		tableName: tableName,
		fieldMap:  fieldMap,
		columnMap: columnMap,
	}

	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

// parseTag 解析标签: 目的是为了可以拿到用户自定义的列名
// 我希望用户是 “orm:"column=id, xxx=xx" 这样子定义列名
func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag, ok := tag.Lookup("orm")
	if !ok {
		return map[string]string{}, nil
	}
	pairs := strings.Split(ormTag, ",")
	res := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		segs := strings.Split(pair, "=")
		if len(segs) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		key := segs[0]
		val := segs[1]
		res[key] = val
	}
	return res, nil
}

// underscoreCase 将驼峰命名转换为下划线分隔的小写形式
func underscoreCase(s string) string {
	// 应用正则转换
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")

	// 转换为小写
	snake = strings.ToLower(snake)

	// 将多个连续的下划线替换为单个下划线
	snake = matchMultipleUnderscores.ReplaceAllString(snake, "_")

	return snake
}
