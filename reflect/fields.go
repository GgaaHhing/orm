package reflect

import (
	"errors"
	"reflect"
)

func IterateFields(entity any) (map[string]any, error) {
	if entity == nil {
		return nil, errors.New("entity 不支持nil")
	}
	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)
	if val.IsZero() {
		return nil, errors.New("不支持零值")
	}

	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, errors.New("entity 不支持类型")
	}

	numField := typ.NumField()
	res := make(map[string]any, numField)
	for i := 0; i < numField; i++ {
		fieldType := typ.Field(i)
		fieldValue := val.Field(i)
		// 判断是否是私有, 私有的字段只能拿到信息，拿不到值
		if fieldType.IsExported() {
			// Interface() 方法让我们能够获取字段的真实值
			res[fieldType.Name] = fieldValue.Interface()
		} else {
			res[fieldType.Name] = reflect.Zero(fieldType.Type).Interface()
		}
	}
	return res, nil
}

func SetField(entity any, field string, newValue any) error {
	// 解析
	val := reflect.ValueOf(entity)
	for val.Type().Kind() == reflect.Pointer {
		// 用于获取指针、接口或其他间接类型指向的实际类型
		val = val.Elem()
	}
	// val可以反向获取字段信息，但是type不能获取val
	fieldVal := val.FieldByName(field)
	if !fieldVal.CanSet() {
		return errors.New("不可修改字段")
	}
	fieldVal.Set(reflect.ValueOf(newValue))
	return nil
}
