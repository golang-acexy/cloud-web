package webcloud

import (
	"fmt"
	"github.com/acexy/golang-toolkit/math/conversion"
)

type Authority[T any] struct {
}

type Pager[T any] struct {
	Records []*T  `json:"records"` // 响应数据
	Total   int64 `json:"total"`   // 响应总记录数

	Size   int `json:"size" form:"size"  binding:"required"`     // 请求每页记录数
	Number int `json:"number" form:"number"  binding:"required"` // 请求页码 从1开始
}

type IDType interface {
	~int | ~uint | ~int32 | ~uint32 | ~int64 | ~uint64 | ~string
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

type BaseBizService[ID IDType, S, M, Q, T any] interface {

	// Save 保存数据
	Save(save map[string]any) (ID, error)

	// QueryByID 通过主键查询
	QueryByID(id ID, result *T) (int64, error)

	// QueryOne 通过条件查询一条数据
	QueryOne(condition map[string]any, result *T) (int64, error)

	// QueryByPager 分页查询
	QueryByPager(condition map[string]any, pager *Pager[T])

	// ModifyByID 通过主键修改数据
	ModifyByID(id ID, update map[string]any) (int64, error)

	// RemoveByID 通过主键删除数据
	RemoveByID(id ID) (int64, error)
}
