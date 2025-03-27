package reflect

import "reflect"

func IterateFunc(entity any) (map[string]FuncInfo, error) {
	typ := reflect.TypeOf(entity)
	numMethod := typ.NumMethod()
	res := make(map[string]FuncInfo, numMethod)
	for i := 0; i < numMethod; i++ {
		method := typ.Method(i)
		fn := method.Func

		// 返回函数类型的输入参数数量
		numIn := fn.Type().NumIn()
		input := make([]reflect.Type, 0, numIn)
		inputValues := make([]reflect.Value, 0, numIn)

		// 结构体方法等价于 方法里面传入的第0个参数是该结构体
		input = append(input, reflect.TypeOf(entity))
		// 所以第0个参数的值也就是 结构体内的值
		inputValues = append(inputValues, reflect.ValueOf(entity))

		for j := 1; j < numIn; j++ {
			// In 返回函数类型的第 i 个输入参数的类型。
			fnInType := fn.Type().In(j)
			input = append(input, fnInType)
			inputValues = append(inputValues, reflect.Zero(fnInType))
		}

		numOut := fn.Type().NumOut()
		output := make([]reflect.Type, 0, numOut)
		for j := 0; j < numOut; j++ {
			output = append(output, fn.Type().Out(j))
		}

		// 使用输入参数 in 调用函数 fn
		// 将输出结果作为值返回
		resValue := fn.Call(inputValues)
		result := make([]any, 0, len(resValue))
		for _, v := range resValue {
			result = append(result, v.Interface())
		}

		res[method.Name] = FuncInfo{
			Name:       method.Name,
			InputType:  input,
			OutputType: output,
			Result:     result,
		}
	}
	return res, nil
}

type FuncInfo struct {
	Name string
	// 下标0指向接收器
	InputType  []reflect.Type
	OutputType []reflect.Type
	Result     []any
}
