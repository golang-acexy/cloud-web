package webcloud

import (
	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/util/coll"
	reflectutil "github.com/acexy/golang-toolkit/util/reflect"
	"github.com/acexy/golang-toolkit/util/str"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

type mode int8

const (
	queryMode mode = iota
	modifyMode
)

// defaultForbiddenColumns excludes IDs and common creation/update columns, but not soft-delete columns.
var defaultForbiddenColumns = []string{
	"id",
	"created_at",
	"create_time",
	"modified_at",
	"modified_time",
	"update_time",
	"update_at",
	"updated_at",
}

// BaseRouter adapts a shared service to CRUD handlers with DTO field allowlists.
// Construct it before serving requests; concurrent safety depends on the service and authority callback.
type BaseRouter[ID IDType, S, M, Q, D any] struct {
	baseBizService BaseBizService[ID, S, M, Q, D]

	authorityFetch     AuthorityFetch[ID]
	authorityValidate  bool
	authorityDataField AuthorityDataField

	modifyAllowedColumns []string
	queryAllowedColumns  []string
}

// BaseRouterHandlers supplies factories for the seven registered CRUD handlers.
// Pass the concrete embedding router to dispatch overridden methods.
type BaseRouterHandlers interface {
	// Save creates a save handler.
	Save() ginstarter.HandlerWrapper
	// QueryByID creates a single-ID query handler.
	QueryByID() ginstarter.HandlerWrapper
	// QueryOne creates a single-result condition handler.
	QueryOne() ginstarter.HandlerWrapper
	// Query creates a list query handler.
	Query() ginstarter.HandlerWrapper
	// QueryPage creates a page query handler.
	QueryPage() ginstarter.HandlerWrapper
	// ModifyByID creates an ID update handler.
	ModifyByID() ginstarter.HandlerWrapper
	// RemoveByID creates an ID delete handler.
	RemoveByID() ginstarter.HandlerWrapper
}

func structNamesToColumns(structName []string) []string {
	return coll.SliceCollect(structName, func(field string) string {
		return structFieldToColumn(field)
	})
}

// NewBaseRouter derives column allowlists from non-pointer Q and M structs.
// It panics on invalid DTO kinds or ambiguous embedded fields. Tags do not define the allowlist.
// Supply a non-nil service and a struct S for save binding; S and the service are not validated here.
func NewBaseRouter[ID IDType, S, M, Q, D any](baseBizService BaseBizService[ID, S, M, Q, D]) *BaseRouter[ID, S, M, Q, D] {
	var q Q
	var m M

	queryFieldNames, err := collectStructFieldNames(q)
	if err != nil {
		panic(err)
	}
	modifyFieldNames, err := collectStructFieldNames(m)
	if err != nil {
		panic(err)
	}
	return &BaseRouter[ID, S, M, Q, D]{
		baseBizService: baseBizService,
		modifyAllowedColumns: coll.SliceFilter(structNamesToColumns(modifyFieldNames), func(field string) bool {
			return !coll.SliceContains(defaultForbiddenColumns, field)
		}),
		queryAllowedColumns: structNamesToColumns(queryFieldNames),
	}
}

// NewBaseRouterWithAuthority adds identity injection to the standard handlers.
// It shares NewBaseRouter's panics; authority configuration is checked only when used.
// Custom handlers and direct service calls must enforce authority themselves.
func NewBaseRouterWithAuthority[ID IDType, S, M, Q, D any](baseBizService BaseBizService[ID, S, M, Q, D], authorityFetch AuthorityFetch[ID], authorityDataField AuthorityDataField) *BaseRouter[ID, S, M, Q, D] {
	router := NewBaseRouter[ID, S, M, Q, D](baseBizService)
	router.authorityFetch = authorityFetch
	router.authorityValidate = true
	router.authorityDataField = authorityDataField
	return router
}

func (b *BaseRouter[ID, S, M, Q, D]) convertJSONToMap(request *ginstarter.Request, m mode) (map[string]any, error) {
	var param map[string]any
	if err := request.BindBodyJSON(&param); err != nil {
		return nil, err
	}
	if len(param) == 0 {
		return nil, ErrBadRequestParameters
	}
	if !b.checkField(param, m) {
		return nil, ErrBadRequestParameters
	}
	return convertFieldsToColumns(param), nil
}

func convertFieldsToColumns(param map[string]any) map[string]any {
	return coll.MapCollect(param, func(k string, v any) (string, any) {
		return str.CamelToSnake(k), v
	})
}

func (b *BaseRouter[ID, S, M, Q, D]) checkField(param map[string]any, m mode) bool {
	var matchRule []string
	switch m {
	case modifyMode:
		matchRule = b.modifyAllowedColumns
	case queryMode:
		matchRule = b.queryAllowedColumns
	}
	input := coll.MapFilterToSlice(param, func(k string, v any) (string, bool) {
		return str.CamelToSnake(k), true
	})
	if !coll.SliceIsSubset(input, matchRule) {
		logger.Logrus().Warningln("some request field not allowed, all request field : ", input)
		return false
	}
	return true
}

func (b *BaseRouter[ID, S, M, Q, D]) normalizeTimeRanges(timeRanges []TimeRange) ([]TimeRange, bool) {
	result, err := normalizeTimeRanges(
		timeRanges,
		b.baseBizService.AllowedTimeRangeFields(),
		b.baseBizService.DefaultTimeRangeField(),
	)
	if err != nil {
		logger.Logrus().Warningln("invalid time ranges")
		return nil, false
	}
	return result, true
}

func (b *BaseRouter[ID, S, M, Q, D]) setAuthorityLimitStruct(request *ginstarter.Request, paramPtr any) error {
	if b.authorityValidate {
		if b.authorityDataField.StructField == "" {
			return ErrAuthorityStructFieldRequired
		}
		authority := b.GetAuthorityData(request)
		err := reflectutil.SetFieldValue(paramPtr, map[string]any{
			b.authorityDataField.StructField: authority.GetIdentityID(),
		}, true)
		if err != nil {
			logger.Logrus().Errorln("set authority field error:", err)
			return err
		}
	}
	return nil
}

func (b *BaseRouter[ID, S, M, Q, D]) setAuthorityLimitMap(request *ginstarter.Request, param map[string]any) error {
	if b.authorityValidate {
		if b.authorityDataField.Column == "" {
			return ErrAuthorityDataColumnRequired
		}
		authority := b.GetAuthorityData(request)
		// Apply authority after client key normalization so client values cannot override it.
		param[b.authorityDataField.Column] = authority.GetIdentityID()
	}
	return nil
}

// RegisterBaseHandlers registers seven relative CRUD routes with JSON-only body routes.
// Call before serving requests with non-nil router and handlers; overrides own their validation.
func (b *BaseRouter[ID, S, M, Q, D]) RegisterBaseHandlers(router *ginstarter.RouterWrapper, handlers BaseRouterHandlers) {
	router.POST1("save", []string{gin.MIMEJSON}, handlers.Save())

	router.GET("by-id/:id", handlers.QueryByID())

	router.POST1("query-one", []string{gin.MIMEJSON}, handlers.QueryOne())

	router.POST1("query", []string{gin.MIMEJSON}, handlers.Query())

	router.POST1("query-by-page", []string{gin.MIMEJSON}, handlers.QueryPage())

	router.PUT1("by-id/:id", []string{gin.MIMEJSON}, handlers.ModifyByID())

	router.DELETE("by-id/:id", handlers.RemoveByID())
}

// GetAuthorityData requires a callback and non-nil authority; otherwise it raises a request panic.
// A typed nil implementation is not detected, and identity values are not validated.
func (b *BaseRouter[ID, S, M, Q, D]) GetAuthorityData(request *ginstarter.Request) Authority[ID] {
	if b.authorityFetch == nil {
		request.Panic(ginstarter.StatusCodeException, ErrAuthorityFetchRequired)
	}
	result := b.authorityFetch(request)
	if result == nil {
		request.Panic(ginstarter.StatusCodeUnauthorized, ErrUnauthorized)
	}
	return result
}

// GetOptionalAuthorityData allows nil authority but raises a request panic if no callback is configured.
func (b *BaseRouter[ID, S, M, Q, D]) GetOptionalAuthorityData(request *ginstarter.Request) Authority[ID] {
	if b.authorityFetch == nil {
		request.Panic(ginstarter.StatusCodeException, ErrAuthorityFetchRequired)
	}
	return b.authorityFetch(request)
}

// Save binds and validates S, then injects authority and calls SaveWithoutZeroFields.
// Binding failures panic through the request error pipeline; validation is not repeated after injection.
// Service failures log the bound DTO; override this handler when payloads require redaction.
func (b *BaseRouter[ID, S, M, Q, D]) Save() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		var param S
		request.MustBindBodyJSON(&param)
		if err := b.setAuthorityLimitStruct(request, &param); err != nil {
			return nil, err
		}
		id, err := b.baseBizService.SaveWithoutZeroFields(&param)
		if err != nil {
			logger.Logrus().Errorln("cant save:", param, err)
			return nil, err
		}
		return ginstarter.RespRestSuccess(id), nil
	}
}

// QueryByID parses the path ID and queries with authority conditions.
// Invalid IDs yield bad parameters; a nil result without error yields success without data.
func (b *BaseRouter[ID, S, M, Q, D]) QueryByID() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		id, err := ConvertStringToID[ID](request.GetPathParam("id"))
		if err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		param := map[string]any{"id": id}
		if err := b.setAuthorityLimitMap(request, param); err != nil {
			return nil, err
		}
		d, err := b.baseBizService.BaseQueryByID(param)
		if err != nil {
			return nil, err
		}
		if d != nil {
			return ginstarter.RespRestSuccess(d), nil
		}
		return ginstarter.RespRestSuccess(), nil
	}
}

// Query accepts a nonempty JSON map with allowed Q columns and injects authority.
// Values are not validated against Q; the service owns list limits and result normalization.
func (b *BaseRouter[ID, S, M, Q, D]) Query() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		param, err := b.convertJSONToMap(request, queryMode)
		if err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		if err := b.setAuthorityLimitMap(request, param); err != nil {
			return nil, err
		}
		ds, err := b.baseBizService.BaseQuery(param)
		if err != nil {
			return nil, err
		}
		return ginstarter.RespRestSuccess(ds), nil
	}
}

// QueryOne validates nonempty condition keys and injects authority before querying.
// A nil result without error yields success without data; Q field validation is not applied.
func (b *BaseRouter[ID, S, M, Q, D]) QueryOne() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		param, err := b.convertJSONToMap(request, queryMode)
		if err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		if err := b.setAuthorityLimitMap(request, param); err != nil {
			return nil, err
		}
		d, err := b.baseBizService.BaseQueryOne(param)
		if err != nil {
			return nil, err
		}
		if d == nil {
			return ginstarter.RespRestSuccess(), nil
		}
		return ginstarter.RespRestSuccess(d), nil
	}
}

// QueryPage validates HTTP page bounds, condition keys, and allowed time intervals.
// Conditions may be omitted; the service receives authority predicates and owns the returned page.
func (b *BaseRouter[ID, S, M, Q, D]) QueryPage() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		var requestParam PagerDTO[map[string]any]
		if err := request.BindBodyJSON(&requestParam); err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		param := requestParam.Condition
		if param == nil {
			param = make(map[string]any)
		}
		if len(param) > 0 {
			if !b.checkField(param, queryMode) {
				return ginstarter.RespRestBadParameters(), nil
			}
			param = convertFieldsToColumns(param)
		}

		if err := b.setAuthorityLimitMap(request, param); err != nil {
			return nil, err
		}
		timeRanges, valid := b.normalizeTimeRanges(requestParam.TimeRanges)
		if !valid {
			return ginstarter.RespRestBadParameters(), nil
		}
		pager, err := b.baseBizService.BaseQueryPage(PagerDTO[map[string]any]{
			Number:     requestParam.Number,
			Size:       requestParam.Size,
			Condition:  param,
			TimeRanges: timeRanges,
		})
		if err != nil {
			return nil, err
		}
		return ginstarter.RespRestSuccess(pager), nil
	}
}

// ModifyByID accepts nonempty allowed M column assignments, preserving zero and null values.
// It enforces authority in both conditions and assignments; zero affected rows yield bad parameters.
// M field types and binding tags are not checked.
func (b *BaseRouter[ID, S, M, Q, D]) ModifyByID() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		id, err := ConvertStringToID[ID](request.GetPathParam("id"))
		if err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		var update map[string]any
		if err = request.BindBodyJSON(&update); err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		if len(update) == 0 {
			return ginstarter.RespRestBadParameters(), nil
		}
		if !b.checkField(update, modifyMode) {
			return ginstarter.RespRestBadParameters(), nil
		}
		update = convertFieldsToColumns(update)
		param := map[string]any{"id": id}
		if err := b.setAuthorityLimitMap(request, param); err != nil {
			return nil, err
		}

		if b.authorityValidate {
			// Enforce the same identity in both the predicate and the assignments.
			update[b.authorityDataField.Column] = param[b.authorityDataField.Column]
		}
		row, err := b.baseBizService.BaseModifyByID(update, param)
		if err != nil {
			return nil, err
		}
		if row == 0 {
			return ginstarter.RespRestBadParameters(), nil
		}
		return ginstarter.RespRestSuccess(), nil
	}
}

// RemoveByID parses the path ID and deletes with authority conditions.
// Nonpositive affected counts yield bad parameters; service errors are forwarded.
func (b *BaseRouter[ID, S, M, Q, D]) RemoveByID() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		id, err := ConvertStringToID[ID](request.GetPathParam("id"))
		if err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		param := map[string]any{"id": id}
		if err := b.setAuthorityLimitMap(request, param); err != nil {
			return nil, err
		}
		row, err := b.baseBizService.BaseRemoveByID(param)
		if err != nil {
			return nil, err
		}
		if row > 0 {
			return ginstarter.RespRestSuccess(), nil
		}
		return ginstarter.RespRestBadParameters(), nil
	}
}

// SimpleRouter exposes identity lookup without automatic route or persistence protection.
type SimpleRouter[ID IDType] struct {
	authorityFetch AuthorityFetch[ID]
}

// NewSimpleRouter stores a callback that must support concurrent requests.
// A nil callback is accepted here but panics through the request pipeline when used.
func NewSimpleRouter[ID IDType](authorityFetch AuthorityFetch[ID]) *SimpleRouter[ID] {
	return &SimpleRouter[ID]{
		authorityFetch: authorityFetch,
	}
}

// GetAuthorityData requires a callback and non-nil authority; otherwise it raises a request panic.
// A typed nil implementation is not detected, and identity values are not validated.
func (s *SimpleRouter[ID]) GetAuthorityData(request *ginstarter.Request) Authority[ID] {
	if s.authorityFetch == nil {
		request.Panic(ginstarter.StatusCodeException, ErrAuthorityFetchRequired)
	}
	result := s.authorityFetch(request)
	if result == nil {
		request.Panic(ginstarter.StatusCodeUnauthorized, ErrUnauthorized)
	}
	return result
}

// GetOptionalAuthorityData allows nil authority but raises a request panic if no callback is configured.
func (s *SimpleRouter[ID]) GetOptionalAuthorityData(request *ginstarter.Request) Authority[ID] {
	if s.authorityFetch == nil {
		request.Panic(ginstarter.StatusCodeException, ErrAuthorityFetchRequired)
	}
	return s.authorityFetch(request)
}
