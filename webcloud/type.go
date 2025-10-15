package webcloud

import (
	"fmt"

	"github.com/acexy/golang-toolkit/math/conversion"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

type Platform string

// IDType 主键类型
type IDType interface {
	~int | ~uint | ~int32 | ~uint32 | ~int64 | ~uint64 | ~string
}

type Authority[ID IDType] interface {
	// GetIdentityID 获取唯一标识
	GetIdentityID() ID
	// GetPlatform 所属平台标识
	GetPlatform() Platform
}

// AuthorityFetch 获取权限信息
type AuthorityFetch[ID IDType] func(request *ginstarter.Request) Authority[ID]

// Pager 分页响应信息
type Pager[T any] struct {
	Records []*T  `json:"records"` // 响应数据
	Total   int64 `json:"total"`   // 响应总记录数

	Size   int `json:"size"` // 请求每页记录数
	Number int `json:"number"`
}

// PagerDTO 分页查询信息
type PagerDTO[T any] struct {
	Size      int `json:"size" form:"size"  binding:"required"`     // 请求每页记录数
	Number    int `json:"number" form:"number"  binding:"required"` // 请求页码 从1开始
	Condition T   `json:"condition"`
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

type BaseBizService[ID IDType, S, M, Q, D any] interface {

	// MaxQueryCount 批量条件查询时，默认最大查询数量
	MaxQueryCount() int

	// DefaultOrderBySQL 默认排序字段
	DefaultOrderBySQL() string

	// Save 保存数据
	Save(save *S) (ID, error)

	// BaseQueryByID 通过主键查询
	BaseQueryByID(condition map[string]any, result *D) (int64, error)

	// BaseQueryOne 通过条件查询一条数据
	BaseQueryOne(condition map[string]any, result *D) (int64, error)

	// BaseQuery 通过条件多条数据
	BaseQuery(condition map[string]any, result *[]*D) (int64, error)

	// BaseQueryByPager 分页查询
	BaseQueryByPager(condition map[string]any, pager *Pager[D]) error

	// BaseModifyByID 通过主键修改数据
	BaseModifyByID(update, condition map[string]any) (int64, error)

	// BaseRemoveByID 通过主键删除数据
	BaseRemoveByID(condition map[string]any) (int64, error)

	// QueryByID 通过主键查询
	QueryByID(id ID) *D

	// QueryOneByCond 通过条件查询一条数据
	QueryOneByCond(condition *Q) *D

	// QueryByCond 通过条件查询多条数据
	QueryByCond(condition *Q) []*D

	// QueryByPager 分页查询
	QueryByPager(pager PagerDTO[Q]) Pager[D]

	// ModifyByID 根据主键修改数据
	ModifyByID(updated *M) bool

	// ModifyByIDExcludeZeroField 根据主键修改数据，忽略值为零的字段
	ModifyByIDExcludeZeroField(updated *M) bool

	// ModifyByIdUseMap 根据主键修改数据
	ModifyByIdUseMap(updated map[string]any, id ID) bool

	// RemoveByID 根据主键删除数据
	RemoveByID(id ID) bool

	// RemoveByCond 根据条件删除数据
	RemoveByCond(condition *D) bool

	// RemoveByMap 根据条件删除数据
	RemoveByMap(condition map[string]any) bool
}
