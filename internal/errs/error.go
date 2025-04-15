package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly   = errors.New("orm：只支持指向结构体的一级指针")
	ErrDeleteALL     = errors.New("orm：不允许直接删除整张表")
	ErrInsertZeroRow = errors.New("orm：插入0行")
	ErrNoRows        = errors.New("orm: 没有数据")
)

func NewErrUnsupportedExpression(expr any) error {
	return fmt.Errorf("orm：不支持的表达式类型 %v", expr)
}

func NewErrUnknownField(name string) error {
	return fmt.Errorf("orm：未知的字段名 %s", name)
}

func NewErrUnknownColumn(name string) error {
	return fmt.Errorf("orm：未知的列名 %s", name)
}

func NewErrInvalidTagContent(pair string) error {
	return fmt.Errorf("orm：非法标签 %s", pair)
}

func NewErrUnsupportedAssignable(assign any) error {
	return fmt.Errorf("orm：不支持的赋值表达式类型 %s", assign)
}
