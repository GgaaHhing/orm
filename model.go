package orm

import (
	"reflect"
	"regexp"
	"strings"
	"web/orm/internal/errs"
)

var (
	matchFirstCap            = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap              = regexp.MustCompile("([a-z0-9])([A-Z])")
	matchMultipleUnderscores = regexp.MustCompile("_+")
)

type model struct {
	tableName string
	fields    map[string]*field
}

type field struct {
	// 列名
	colName string
}

//var defaultRegistry = &registry{
//	models: map[reflect.Type]*model{},
//}

// registry 元数据注册中心
type registry struct {
	models map[reflect.Type]*model
}

func NewRegistry() *registry {
	return &registry{
		models: make(map[reflect.Type]*model, 64),
	}
}

func (r *registry) get(val any) (*model, error) {
	typ := reflect.TypeOf(val)
	m, ok := r.models[typ]
	if !ok {
		var err error
		// 如果不ok，说明是我没解析过的，我解析一下
		m, err = r.parseModel(val)
		if err != nil {
			return nil, err
		}
		r.models[typ] = m
	}
	return m, nil
}

func (r *registry) parseModel(entity any) (*model, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()
	numField := typ.NumField()
	fields := make(map[string]*field, numField)
	for i := 0; i < numField; i++ {
		f := typ.Field(i)
		fields[f.Name] = &field{
			colName: underscoreCase(f.Name),
		}
	}
	return &model{
		tableName: underscoreCase(typ.Name()),
		fields:    fields,
	}, nil
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
