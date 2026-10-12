package webcloud

import (
	"reflect"
	"strconv"

	"github.com/acexy/golang-toolkit/util/json"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

// Platform identifies an application-defined authority domain; routers do not enforce it.
type Platform string

// IDType accepts supported integer and string underlying types, including named types.
type IDType interface {
	~int | ~uint | ~int32 | ~uint32 | ~int64 | ~uint64 | ~string
}

// Authority supplies request identity; return a non-nil implementation for protected operations.
type Authority[ID IDType] interface {

	// GetIdentityID returns the identity injected into protected operations; zero values are not rejected.
	GetIdentityID() ID

	// GetPlatform returns application metadata; base routers do not consult it.
	GetPlatform() Platform
}

// AuthorityFetch resolves identity per request; nil means unauthenticated. Implementations must support concurrent requests.
type AuthorityFetch[ID IDType] func(request *ginstarter.Request) Authority[ID]

// AuthorityDataField maps the identity to a save DTO field and a persistence condition column.
type AuthorityDataField struct {
	// StructField names a writable save DTO field; use the identity type to avoid lossy conversion.
	StructField string
	// Column is used verbatim for persistence conditions and protected updates.
	// Use a distinct ownership column when overwriting the resource ID would be incorrect.
	Column string
}

// TimeRange describes a left-closed, right-open interval validated by the router or ConvertTimeRanges.
type TimeRange struct {
	// Field is an allowed camelCase or snake_case name; empty selects the default field.
	Field string `json:"field"`
	// Start is the inclusive lower bound; nil leaves it unbounded.
	Start *json.Timestamp `json:"start"`
	// End is the exclusive upper bound; nil leaves it unbounded.
	End *json.Timestamp `json:"end"`
}

// PagerDTO holds page inputs; binding tags apply to HTTP requests, not direct service calls.
type PagerDTO[T any] struct {
	// Size is the requested page size; HTTP input must be between 1 and 2000.
	Size int `json:"size" form:"size"  binding:"required,gte=1,lte=2000"`
	// Number is the one-based page number.
	Number int `json:"number" form:"number"  binding:"required,gte=1"`
	// Condition holds service-specific filters; HTTP map conditions may be omitted.
	Condition T `json:"condition"`
	// TimeRanges holds optional intervals; at most one may omit its field.
	TimeRanges []TimeRange `json:"timeRanges"`
}

// Pager holds service-produced page output; the router does not normalize records or metadata.
type Pager[T any] struct {
	// Records contains the returned records; nil and empty slices retain their encoding semantics.
	Records []*T `json:"records"`
	// Total is the service-reported total matching count.
	Total int64 `json:"total"`
	// Size is the service-reported page size; the router does not clamp it.
	Size int `json:"size"`
	// Number is the one-based page number.
	Number int `json:"number"`
}

// ConvertStringToID parses decimal integers with overflow checks or preserves strings verbatim.
// It accepts zero, signed negative IDs, and empty strings; domain validation belongs to the caller.
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

// BaseBizService defines application-owned persistence operations for shared HTTP routers.
// Implementations must support concurrent requests, preserve authority conditions, and return persistence errors.
// List limits, sorting, missing-result conventions, transactions, and cancellation are implementation responsibilities.
type BaseBizService[ID IDType, S, M, Q, D any] interface {

	// MaxQuerySize reports the service list limit; the router does not enforce it.
	MaxQuerySize() int

	// DefaultOrderBy reports a persistence-specific default ordering for the service to apply.
	DefaultOrderBy() string

	// DefaultTimeRangeField supplies the field for one unnamed time range; it must also be allowed.
	DefaultTimeRangeField() string

	// AllowedTimeRangeFields lists permitted time fields, normalized to snake case before comparison.
	AllowedTimeRangeFields() []string

	// Save persists a save DTO and returns its ID; zero-value behavior depends on the implementation.
	Save(save *S) (ID, error)

	// SaveWithoutZeroFields persists a save DTO while omitting zero fields; used by the base save route.
	SaveWithoutZeroFields(save *S) (ID, error)

	// SaveBatch persists save DTOs; atomicity and partial-failure behavior depend on the implementation.
	SaveBatch(saves []*S) ([]ID, error)

	// BaseQueryByID queries by column conditions including ID and authority; return nil, nil for a no-data HTTP response.
	BaseQueryByID(condition map[string]any) (*D, error)

	// BaseQueryOne queries one result by column conditions; return nil, nil for a no-data HTTP response.
	BaseQueryOne(condition map[string]any) (*D, error)

	// BaseQuery queries a list by column conditions; the service owns limits, ordering, and nil-versus-empty results.
	BaseQuery(condition map[string]any) ([]*D, error)

	// BaseQueryPage queries a page using column conditions and normalized time ranges; preserve all authority predicates.
	BaseQueryPage(query PagerDTO[map[string]any]) (Pager[D], error)

	// BaseModifyByID updates by ID and authority conditions; zero affected rows become a bad-parameters response.
	BaseModifyByID(update, condition map[string]any) (int64, error)

	// BaseRemoveByID deletes by ID and authority conditions; nonpositive affected rows become a bad-parameters response.
	BaseRemoveByID(condition map[string]any) (int64, error)

	// QueryByID queries one ID using the implementation's missing-result convention.
	QueryByID(id ID) (*D, error)

	// QueryByIDs queries multiple IDs; ordering and empty-input behavior depend on the implementation.
	QueryByIDs(ids []ID) ([]*D, error)

	// ExistsByID checks an ID and returns persistence failures separately from absence.
	ExistsByID(id ID) (bool, error)

	// QueryOneByCond queries one result using typed conditions.
	QueryOneByCond(condition Q) (*D, error)

	// QueryByCond queries a list using typed conditions and service-owned limits.
	QueryByCond(condition Q) ([]*D, error)

	// CountByCond counts results matching typed conditions.
	CountByCond(condition Q) (int64, error)

	// CountByMap counts results using column conditions, including explicit zero values.
	CountByMap(condition map[string]any) (int64, error)

	// QueryPage queries typed page inputs; direct calls must validate their own page limits.
	QueryPage(pager PagerDTO[Q]) (Pager[D], error)

	// ModifyByID updates an ID using a modify DTO; zero-value behavior depends on the implementation.
	ModifyByID(id ID, updated *M) (int64, error)

	// ModifyByIDWithoutZeroFields updates an ID while omitting zero-valued DTO fields.
	ModifyByIDWithoutZeroFields(id ID, updated *M) (int64, error)

	// ModifyByIDWithMap updates an ID using column assignments, including explicit zero values.
	ModifyByIDWithMap(id ID, updated map[string]any) (int64, error)

	// ModifyByCond updates matching typed conditions using a modify DTO.
	ModifyByCond(condition Q, updated *M) (int64, error)

	// ModifyByMap updates column conditions with explicit column assignments, including zero values.
	ModifyByMap(updated, condition map[string]any) (int64, error)

	// RemoveByID deletes one ID and returns the persistence-reported affected count.
	RemoveByID(id ID) (int64, error)

	// RemoveByIDs deletes multiple IDs; empty-input behavior depends on the implementation.
	RemoveByIDs(ids []ID) (int64, error)

	// RemoveByCond deletes matching typed conditions; the implementation owns empty-filter safeguards.
	RemoveByCond(condition Q) (int64, error)

	// RemoveByMap deletes matching column conditions; the implementation owns empty-filter safeguards.
	RemoveByMap(condition map[string]any) (int64, error)
}
