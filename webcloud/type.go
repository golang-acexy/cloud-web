package webcloud

import (
	"fmt"
	"github.com/acexy/golang-toolkit/math/conversion"
)

type IDType interface {
	~int | ~int32 | ~int64 | ~uint64 | ~uint32 | ~uint | ~string
}

// CovertStringToID 字符串类型转换为实际主键类型
func CovertStringToID[ID IDType](value string) (ID, error) {
	var id ID
	var err error
	var v any
	switch any(id).(type) {
	case int:
		v, err = conversion.ParseIntError(value)
	case int32:
		v, err = conversion.ParseInt32Error(value)
	case int64:
		v, err = conversion.ParseInt64Error(value)
	case uint:
		v, err = conversion.ParseUintError(value)
	case uint32:
		v, err = conversion.ParseUint32Error(value)
	case uint64:
		v, err = conversion.ParseUint64Error(value)
	case string:
		v = value
	default:
		return id, fmt.Errorf("unsupported id type")
	}
	return v.(ID), err
}

type Authority[T any] struct {
}

type BaseBizService[T any, ID IDType] interface {

	// Save 保存数据
	Save(t *T) (ID, error)

	// QueryById 通过主键查询
	QueryById(id ID, result *T) (int64, error)

	// ModifyById 通过主键修改数据
	ModifyById(id ID, update map[string]any) (int64, error)
}
