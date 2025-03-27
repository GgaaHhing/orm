package reflect

import "reflect"

func IterateArrayOrSlice(entity any) ([]any, error) {
	val := reflect.ValueOf(entity)
	res := make([]any, 0, val.Len())
	// 获取长度
	for i := 0; i < val.Len(); i++ {
		ele := val.Index(i)
		res = append(res, ele.Interface())
	}
	return res, nil
}

// IterateMap 第一个参数：key，第二个参数：value
func IterateMap(entity any) ([]any, []any, error) {
	val := reflect.ValueOf(entity)
	resKeys := make([]any, 0, val.Len())
	resVals := make([]any, 0, val.Len())
	keys := val.MapKeys()
	// 获取长度
	for _, key := range keys {
		v := val.MapIndex(key)
		resKeys = append(resKeys, key.Interface())
		resVals = append(resVals, v.Interface())
	}
	return resKeys, resVals, nil
}
