package webcloud

import (
	"reflect"
	"strconv"

	"github.com/acexy/golang-toolkit/util/json"
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

// AuthorityDataField 定义数据权限字段在 Go 结构体和数据库中的名称。
// StructField 用于向请求 DTO 注入权限值，Column 用于构造数据库条件。
type AuthorityDataField struct {
	StructField string
	Column      string
}

// TimeRange 定义某个白名单时间字段的左闭右开查询范围。
type TimeRange struct {
	Field string          `json:"field"` // 支持驼峰或下划线字段名，字段必须位于时间范围白名单中
	Start *json.Timestamp `json:"start"`
	End   *json.Timestamp `json:"end"`
}

// PagerDTO 分页查询信息
type PagerDTO[T any] struct {
	Size       int         `json:"size" form:"size"  binding:"required,gte=1,lte=2000"` // 请求每页记录数
	Number     int         `json:"number" form:"number"  binding:"required,gte=1"`      // 请求页码 从1开始
	Condition  T           `json:"condition"`
	TimeRanges []TimeRange `json:"timeRanges"`
}

// Pager 分页响应信息
type Pager[T any] struct {
	Records []*T  `json:"records"` // 响应数据
	Total   int64 `json:"total"`   // 响应总记录数
	Size    int   `json:"size"`    // 请求每页记录数
	Number  int   `json:"number"`
}

// ConvertStringToID 将字符串转换为实际主键类型。
func ConvertStringToID[ID IDType](value string) (ID, error) {
	var id ID
	targetType := reflect.TypeOf(id)
	var parsedValue reflect.Value
	switch targetType.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(value, 10, targetType.Bits())
		if err != nil {
			return id, err
		}
		parsedValue = reflect.ValueOf(parsed).Convert(targetType)
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(value, 10, targetType.Bits())
		if err != nil {
			return id, err
		}
		parsedValue = reflect.ValueOf(parsed).Convert(targetType)
	case reflect.String:
		parsedValue = reflect.ValueOf(value).Convert(targetType)
	default:
		return id, ErrUnsupportedIDType
	}
	return parsedValue.Interface().(ID), nil
}

type BaseBizService[ID IDType, S, M, Q, D any] interface {

	// MaxQuerySize 非分页查询允许返回的最大记录数。
	MaxQuerySize() int

	// DefaultOrderBy 返回默认排序表达式，具体格式由持久化实现解释。
	DefaultOrderBy() string

	// DefaultTimeRangeField 默认查询时间范围字段。
	DefaultTimeRangeField() string

	// AllowedTimeRangeFields 允许使用的时间范围字段。
	AllowedTimeRangeFields() []string

	// Save 保存数据
	Save(save *S) (ID, error)

	// SaveWithoutZeroFields 保存数据，忽略零值字段。
	SaveWithoutZeroFields(save *S) (ID, error)

	// SaveBatch 批量保存数据。
	SaveBatch(saves []*S) ([]ID, error)

	// BaseQueryByID 使用数据库字段条件查询主键记录。
	BaseQueryByID(condition map[string]any) (*D, error)

	// BaseQueryOne 使用数据库字段条件查询一条记录。
	BaseQueryOne(condition map[string]any) (*D, error)

	// BaseQuery 使用数据库字段条件查询多条记录。
	BaseQuery(condition map[string]any) ([]*D, error)

	// BaseQueryPage 使用数据库字段条件分页查询。
	BaseQueryPage(query PagerDTO[map[string]any]) (Pager[D], error)

	// BaseModifyByID 通过主键修改数据
	BaseModifyByID(update, condition map[string]any) (int64, error)

	// BaseRemoveByID 通过主键删除数据
	BaseRemoveByID(condition map[string]any) (int64, error)

	// QueryByID 通过主键查询。
	QueryByID(id ID) (*D, error)

	// QueryByIDs 通过多个主键查询。
	QueryByIDs(ids []ID) ([]*D, error)

	// ExistsByID 判断指定主键的数据是否存在。
	ExistsByID(id ID) (bool, error)

	// QueryOneByCond 通过条件查询一条数据。
	QueryOneByCond(condition Q) (*D, error)

	// QueryByCond 通过条件查询多条数据。
	QueryByCond(condition Q) ([]*D, error)

	// CountByCond 统计符合条件的数据数量。
	CountByCond(condition Q) (int64, error)

	// CountByMap 根据数据库字段条件统计数据，支持显式零值条件。
	CountByMap(condition map[string]any) (int64, error)

	// QueryPage 分页查询。
	QueryPage(pager PagerDTO[Q]) (Pager[D], error)

	// ModifyByID 根据主键修改数据。
	ModifyByID(id ID, updated *M) (int64, error)

	// ModifyByIDWithoutZeroFields 根据主键修改数据，忽略零值字段。
	ModifyByIDWithoutZeroFields(id ID, updated *M) (int64, error)

	// ModifyByIDWithMap 根据主键使用 Map 修改数据。
	ModifyByIDWithMap(id ID, updated map[string]any) (int64, error)

	// ModifyByCond 根据类型化条件修改数据。
	ModifyByCond(condition Q, updated *M) (int64, error)

	// ModifyByMap 根据数据库字段条件和字段值修改数据，支持显式零值。
	ModifyByMap(updated, condition map[string]any) (int64, error)

	// RemoveByID 根据主键删除数据。
	RemoveByID(id ID) (int64, error)

	// RemoveByIDs 根据多个主键删除数据。
	RemoveByIDs(ids []ID) (int64, error)

	// RemoveByCond 根据查询条件删除数据。
	RemoveByCond(condition Q) (int64, error)

	// RemoveByMap 根据数据库字段条件删除数据。
	RemoveByMap(condition map[string]any) (int64, error)
}
