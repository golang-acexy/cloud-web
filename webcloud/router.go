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

// defaultForbiddenColumns 定义基础更新接口禁止修改的系统字段。
var defaultForbiddenColumns = []string{
	"id",
	"created_at",
	"create_time",
	"modified_at",
	"modified_time",
	"update_time",
	"update_at",
	"updated_at",
	"deleted_at",
}

type BaseRouter[ID IDType, S, M, Q, D any] struct {
	baseBizService BaseBizService[ID, S, M, Q, D]

	// 权限控制
	authorityFetch     AuthorityFetch[ID]
	authorityValidate  bool
	authorityDataField AuthorityDataField

	// 字段安全设置
	modifyAllowedColumns []string // 允许自由更新的数据库字段
	queryAllowedColumns  []string // 允许自由查询的数据库字段
}

func structNamesToColumns(structName []string) []string {
	return coll.SliceCollect(structName, func(field string) string {
		return structFieldToColumn(field)
	})
}

// NewBaseRouter 创建基础路由。
// 查询和更新请求以 Map 形式向 Service 传递，以保留零值字段；允许字段由 Q、M 结构体定义。
// S、M、Q 必须使用非指针结构体类型。
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

// NewBaseRouterWithAuthority 创建基础路由 自动携带数据权限控制
// authorityFetch 提供获取授权信息的接口
// authorityDataField 分别定义权限字段对应的 Go 结构体字段名和数据库列名。
func NewBaseRouterWithAuthority[ID IDType, S, M, Q, D any](baseBizService BaseBizService[ID, S, M, Q, D], authorityFetch AuthorityFetch[ID], authorityDataField AuthorityDataField) *BaseRouter[ID, S, M, Q, D] {
	router := NewBaseRouter[ID, S, M, Q, D](baseBizService)
	router.authorityFetch = authorityFetch
	router.authorityValidate = true
	router.authorityDataField = authorityDataField
	return router
}

// convertJSONToMap 将 JSON 转换成数据库字段 Map，并检查请求字段是否允许。
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

// convertFieldsToColumns 将客户端字段名统一转换为数据库列名。
func convertFieldsToColumns(param map[string]any) map[string]any {
	return coll.MapCollect(param, func(k string, v any) (string, any) {
		return str.CamelToSnake(k), v
	})
}

// checkField 安全检查
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

// setAuthorityLimitStruct 向请求 DTO 强制写入数据权限字段。
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

// setAuthorityLimitMap 向数据库条件强制写入数据权限字段。
func (b *BaseRouter[ID, S, M, Q, D]) setAuthorityLimitMap(request *ginstarter.Request, param map[string]any) error {
	if b.authorityValidate {
		if b.authorityDataField.Column == "" {
			return ErrAuthorityDataColumnRequired
		}
		authority := b.GetAuthorityData(request)
		param[b.authorityDataField.Column] = authority.GetIdentityID()
	}
	return nil
}

// RegisterBaseHandlers 注册基础路由。
func (b *BaseRouter[ID, S, M, Q, D]) RegisterBaseHandlers(router *ginstarter.RouterWrapper) {
	router.POST1("save", []string{gin.MIMEJSON}, b.Save())
	// 通过主键查询单条数据
	router.GET("by-id/:id", b.QueryByID())
	// 通过条件查询单条数据
	router.POST1("query-one", []string{gin.MIMEJSON}, b.QueryOne())
	// 通过条件查询多条数据
	router.POST1("query", []string{gin.MIMEJSON}, b.Query())
	// 通过条件分页查询
	router.POST1("query-by-page", []string{gin.MIMEJSON}, b.QueryPage())
	// 通过主键更新数据
	router.PUT1("by-id/:id", []string{gin.MIMEJSON}, b.ModifyByID())
	// 通过主键删除数据
	router.DELETE("by-id/:id", b.RemoveByID())
}

// GetAuthorityData 获取当前请求的必需认证信息。
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

// GetOptionalAuthorityData 获取当前请求的可选认证信息。
func (b *BaseRouter[ID, S, M, Q, D]) GetOptionalAuthorityData(request *ginstarter.Request) Authority[ID] {
	if b.authorityFetch == nil {
		request.Panic(ginstarter.StatusCodeException, ErrAuthorityFetchRequired)
	}
	return b.authorityFetch(request)
}

// 基础 CRUD

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

func (b *BaseRouter[ID, S, M, Q, D]) QueryPage() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		var requestParam PagerDTO[map[string]any]
		if err := request.BindBodyJSON(&requestParam); err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		pager := Pager[D]{
			Records: make([]*D, 0),
			Number:  requestParam.Number,
			Size:    requestParam.Size,
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
		// 权限条件必须最后写入，确保客户端传值无法覆盖系统权限。
		if err := b.setAuthorityLimitMap(request, param); err != nil {
			return nil, err
		}
		if err := b.baseBizService.BaseQueryPage(param, &pager); err != nil {
			return nil, err
		}
		return ginstarter.RespRestSuccess(pager), nil
	}
}

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
		// 权限字段由系统强制写入，客户端无法修改为其他身份。
		if b.authorityValidate {
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

// SimpleRouter 简单路由，不含数据库结构相关的方法
type SimpleRouter[ID IDType] struct {
	authorityFetch AuthorityFetch[ID]
}

// NewSimpleRouter 创建一个简单路由 该路由器仅含有快捷获取当前认证信息
func NewSimpleRouter[ID IDType](authorityFetch AuthorityFetch[ID]) *SimpleRouter[ID] {
	return &SimpleRouter[ID]{
		authorityFetch: authorityFetch,
	}
}

// GetAuthorityData 获取当前请求的必需认证信息。
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

// GetOptionalAuthorityData 获取当前请求的可选认证信息。
func (s *SimpleRouter[ID]) GetOptionalAuthorityData(request *ginstarter.Request) Authority[ID] {
	if s.authorityFetch == nil {
		request.Panic(ginstarter.StatusCodeException, ErrAuthorityFetchRequired)
	}
	return s.authorityFetch(request)
}
