package unsafe

import (
	"fmt"
	"reflect"
	"unsafe"
)

// UnsafeAccessor
/*
	Go会遵循字长对齐原则，也就是
	String是16字节，int32是4字节，int64是8字节

	为什么unsafe比反射性能更高，因为reflect可以看作是对于unsafe的一个封装
	直接操作unsafe减少了这些封装的开销
*/
type UnsafeAccessor struct {
	fields  map[string]FieldMeta
	address unsafe.Pointer
}

func NewUnsafeAccessor(entity any) *UnsafeAccessor {
	typ := reflect.TypeOf(entity).Elem()
	numFields := typ.NumField()
	fields := make(map[string]FieldMeta, numFields)
	for i := 0; i < numFields; i++ {
		fd := typ.Field(i)
		fields[fd.Name] = FieldMeta{
			Offset: fd.Offset,
			typ:    fd.Type,
		}
	}
	val := reflect.ValueOf(entity)
	return &UnsafeAccessor{
		fields: fields,
		// 存储起始地址
		//
		// address: val.UnsafeAddr()
		// 不用这个是因为，unsafeAddr中的地址是不稳定的
		address: val.UnsafePointer(),
	}
}

// Field 读取字段
// 读：*(T)(ptr)，T是目标类型，如果类型不
// 知道，只能拿到反射的 Type，那么可以用
// reflect. NewAt(typ, ptr).Elem0。
func (a *UnsafeAccessor) Field(field string) (any, error) {
	fd, ok := a.fields[field]
	if !ok {
		return nil, fmt.Errorf("非法字段，field %s not found", field)
	}
	// 计算偏移量
	//fdAddr := uintptr(a.address) + fd.Offset
	// GC 会帮你维护unsafe.Pointer的指针,GO层面的指针
	// uintptr 则是指一段内存地址，就是一个数字
	// 只在计算地址偏移量的时候，采用uintptr
	fdAddr := unsafe.Pointer(uintptr(a.address) + fd.Offset)
	// 之所以用下面的直接加unsafe的写法，而不用上面
	// 是因为，可能在执行下一句return的中间，GC就回收了
	// 如果是int类型
	//return *(*int)(fdAddr), nil

	// 如果不知道类型：
	// 返回一个表示指向指定类型的值的指针的值,使用 p 作为该指针。
	// New和NewAt都会返回指针
	return reflect.NewAt(fd.typ, fdAddr).Elem().Interface(), nil
}

// SetField 修改字段值
// 写：*(AT(ptr)=T，T是目标类型。
func (a *UnsafeAccessor) SetField(field string, value any) error {
	fd, ok := a.fields[field]
	if !ok {
		return fmt.Errorf("非法字段，field %s not found", field)
	}
	// 计算偏移量
	//fdAddr := uintptr(a.address) + fd.Offset
	fdAddr := unsafe.Pointer(uintptr(a.address) + fd.Offset)
	// 如果知道确切类型
	// *(*int)(fdAddr) = val
	reflect.NewAt(fd.typ, fdAddr).Elem().Set(reflect.ValueOf(value))
	return nil
}

type FieldMeta struct {
	Offset uintptr
	typ    reflect.Type
}
