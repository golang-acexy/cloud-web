package webcloud

import (
	"errors"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/acexy/golang-toolkit/util/reflect"
	"github.com/acexy/golang-toolkit/util/str"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

type mode int8

const (
	query mode = iota
	modify
	save
)

// 以下常用的字段用于安全设置，在写场景下强制自动忽略
var defaultForbitColumns = []string{
	"id",
	"created_at",
	"create_time",
	"modified_at",
	"modified_time",
	"update_time",
	"update_at",
}

type BaseRouter[ID IDType, S, M, Q, D any] struct {
	baseBizService BaseBizService[ID, S, M, Q, D]

	// 权限控制
	authorityFetch           AuthorityFetch[ID]
	authorityValidate        bool
	authorityDataLimitColumn string // 权限数据控制的数据库字段

	// 字段安全设置
	modifyAllowedColumns []string // 允许自由更新的数据库字段
	queryAllowedColumns  []string // 允许自由查询的数据库字段
	saveAllowedColumns   []string // 允许自由保存的数据库字段
}

func structNames2Columns(structName []string) []string {
	return coll.SliceCollect(structName, func(field string) string {
		if field == "ID" || field == "Id" {
			return "id"
		}
		return str.CamelToSnake(str.LowFirstChar(field))
	})
}

// NewBaseRouter 创建基础路由
// 为解决零值保存/新增/查询的问题，基础Router的默认动作均将请求参数通过转换为map[jsonKey]jsonValue的形式向后传递参数
// 但这样会增加请求方自由度，出现恶意的请求编辑的jsonKey字段作为数据库直接交互的字段，基础Router根据不同的结构体所具有的字段来限定
// **不要将不能更新的字段用于设置在结构体中**
func NewBaseRouter[ID IDType, S, M, Q, D any](baseBizService BaseBizService[ID, S, M, Q, D]) *BaseRouter[ID, S, M, Q, D] {
	var q Q
	var m M
	var s S

	queryFieldNames, err := reflect.AllFieldName(q)
	if err != nil {
		panic(err)
	}
	modifyFieldNames, err := reflect.AllFieldName(m)
	if err != nil {
		panic(err)
	}
	saveFieldNames, err := reflect.AllFieldName(s)
	if err != nil {
		panic(err)
	}

	return &BaseRouter[ID, S, M, Q, D]{
		baseBizService: baseBizService,
		modifyAllowedColumns: coll.SliceFilter(structNames2Columns(modifyFieldNames), func(field string) bool {
			return !coll.SliceContains(defaultForbitColumns, field)
		}),
		queryAllowedColumns: structNames2Columns(queryFieldNames),
		saveAllowedColumns: coll.SliceFilter(structNames2Columns(saveFieldNames), func(field string) bool {
			return !coll.SliceContains(defaultForbitColumns, field)
		}),
	}
}

// NewBaseRouterWithAuthority 创建基础路由 自动携带数据权限控制
// authorityFetch 提供获取授权信息的接口
// authorityDataLimitFieldName 权限控制的字段名称
func NewBaseRouterWithAuthority[ID IDType, S, M, Q, D any](baseBizService BaseBizService[ID, S, M, Q, D], authorityFetch AuthorityFetch[ID], authorityDataLimitFieldName string) *BaseRouter[ID, S, M, Q, D] {
	router := NewBaseRouter[ID, S, M, Q, D](baseBizService)
	router.authorityFetch = authorityFetch
	router.authorityValidate = true
	router.authorityDataLimitColumn = str.CamelToSnake(str.LowFirstChar(authorityDataLimitFieldName))
	if !coll.SliceContains(router.queryAllowedColumns, authorityDataLimitFieldName) {
		router.queryAllowedColumns = append(router.queryAllowedColumns, authorityDataLimitFieldName)
	}
	return router
}

// ConvertJsonToMap 将json转换成map
// 同时检查请求的字段是否允许 注意，key为自动转换成数据库字段名
func (b *BaseRouter[ID, S, M, Q, D]) ConvertJsonToMap(request *ginstarter.Request, m mode) (map[string]any, error) {
	var param map[string]any
	rawBytes, err := request.GetRawBodyData()
	if err != nil {
		return nil, err
	}
	json.ParseBytesPanic(rawBytes, &param)
	if len(param) == 0 {
		return nil, errors.New("bad request param")
	}
	if !b.checkField(param, m) {
		return nil, errors.New("bad request param")
	}
	return coll.MapCollect(param, func(k string, v any) (string, any) {
		return str.CamelToSnake(k), v
	}), nil
}

// checkField 安全检查
func (b *BaseRouter[ID, S, M, Q, D]) checkField(param map[string]any, m mode) bool {
	var mathRule []string
	switch m {
	case save:
		mathRule = b.saveAllowedColumns
	case modify:
		mathRule = b.modifyAllowedColumns
	case query:
		mathRule = b.queryAllowedColumns
	}
	input := coll.MapFilterToSlice(param, func(k string, v any) (string, bool) {
		return str.CamelToSnake(k), true
	})
	if !coll.SliceIsSubset(input, mathRule) {
		logger.Logrus().Warningln("some request field not allowed, all request field : ", input)
		return false
	}
	return true
}

// SetAuthorityLimitStruct 针对于需要数据权限控制的路由，设置数据权限控制字段
func (b *BaseRouter[ID, S, M, Q, D]) SetAuthorityLimitStruct(request *ginstarter.Request, paramPtr any) (bool, error) {
	if b.authorityValidate {
		authority := b.GetAuthorityData(request)
		if authority == nil {
			return false, nil
		}
		err := reflect.SetFieldValue(paramPtr, map[string]any{
			b.authorityDataLimitColumn: authority.GetIdentityID(),
		})
		if err != nil {
			logger.Logrus().Errorln("set authority field error:", err)
			return false, err
		}
	}
	return true, nil
}

// SetAuthorityLimitMap 针对于需要数据权限控制的路由，设置数据权限控制字段
// 需传入数据库字段名
func (b *BaseRouter[ID, S, M, Q, D]) SetAuthorityLimitMap(request *ginstarter.Request, param map[string]any) bool {
	if b.authorityValidate {
		authority := b.GetAuthorityData(request)
		if authority == nil {
			return false
		}
		param[str.CamelToSnake(b.authorityDataLimitColumn)] = authority.GetIdentityID()
	}
	return true
}

// RegisterBaseHandler 注册基础路由
func (b *BaseRouter[ID, S, M, Q, D]) RegisterBaseHandler(router *ginstarter.RouterWrapper, baseRouter *BaseRouter[ID, S, M, Q, D]) {
	router.POST1("save", []string{gin.MIMEJSON}, baseRouter.Save())
	// 通过主键查询单条数据
	router.GET("by-id/:id", baseRouter.QueryById())
	// 通过条件查询单条数据
	router.POST1("query-one", []string{gin.MIMEJSON}, baseRouter.QueryOne())
	// 通过条件查询多条数据
	router.POST1("query", []string{gin.MIMEJSON}, baseRouter.Query())
	// 通过条件分页查询
	router.POST1("query-by-page", []string{gin.MIMEJSON}, baseRouter.QueryByPage())
	// 通过主键更新数据
	router.PUT1("by-id/:id", []string{gin.MIMEJSON}, baseRouter.ModifyById())
	// 通过主键删除数据
	router.DELETE("by-id/:id", baseRouter.RemoveById())
}

// GetAuthorityData 获取当前请求的认证信息
func (b *BaseRouter[ID, S, M, Q, D]) GetAuthorityData(request *ginstarter.Request, notMust ...bool) Authority[ID] {
	if b.authorityFetch == nil {
		logger.Logrus().Warningln("not set authority fetch method")
		return nil
	}
	result := b.authorityFetch(request)
	if len(notMust) > 0 && notMust[0] {
		return result
	}
	if result == nil {
		request.Panic(ginstarter.StatusCodeUnauthorized, errors.New("Unauthorized Request"))
	}
	return b.authorityFetch(request)
}

// 基础CRUD

func (b *BaseRouter[ID, S, M, Q, D]) Save() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		var param S
		request.MustBindBodyAuto(&param)
		pass, err := b.SetAuthorityLimitStruct(request, param)
		if err != nil {
			return nil, err
		}
		if !pass {
			return ginstarter.RespRestUnAuthorized(), nil
		}
		id, err := b.baseBizService.Save(&param)
		if err != nil {
			logger.Logrus().Errorln("cant save:", param, err)
			return nil, err
		}
		return ginstarter.RespRestSuccess(id), nil
	}
}

func (b *BaseRouter[ID, S, M, Q, D]) QueryById() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		id, err := CovertStringToID[ID](request.GetPathParam("id"))
		if err != nil {
			return nil, err
		}
		param := map[string]any{"id": id}
		flag := b.SetAuthorityLimitMap(request, param)
		if !flag {
			return ginstarter.RespRestUnAuthorized(), nil
		}
		var d D
		row, err := b.baseBizService.BaseQueryByID(param, &d)
		if err != nil {
			return nil, err
		}
		if row > 0 {
			return ginstarter.RespRestSuccess(d), nil
		}
		return ginstarter.RespRestSuccess(), nil
	}
}

func (b *BaseRouter[ID, S, M, Q, D]) Query() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		param, err := b.ConvertJsonToMap(request, query)
		if err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		flag := b.SetAuthorityLimitMap(request, param)
		if !flag {
			return ginstarter.RespRestUnAuthorized(), nil
		}
		var ds []*D
		row, err := b.baseBizService.BaseQuery(param, &ds)
		if err != nil {
			return nil, err
		}
		if row == 0 {
			return ginstarter.RespRestSuccess(), nil
		}
		return ginstarter.RespRestSuccess(ds), nil
	}
}

func (b *BaseRouter[ID, S, M, Q, D]) QueryOne() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		param, err := b.ConvertJsonToMap(request, query)
		if err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		flag := b.SetAuthorityLimitMap(request, param)
		if !flag {
			return ginstarter.RespRestUnAuthorized(), nil
		}
		var d D
		row, err := b.baseBizService.BaseQueryOne(param, &d)
		if err != nil {
			return nil, err
		}
		if row == 0 {
			return ginstarter.RespRestSuccess(), nil
		}
		return ginstarter.RespRestSuccess(d), nil
	}
}

func (b *BaseRouter[ID, S, M, Q, D]) QueryByPage() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		rawBytes, err := request.GetRawBodyData()
		if err != nil {
			return nil, err
		}
		paramJson := json.NewGJsonBytes(rawBytes)
		sizeValue := paramJson.Get("size")
		size, exist := sizeValue.IntValue()
		if !exist {
			return ginstarter.RespRestBadParameters("size required"), nil
		}
		numberValue := paramJson.Get("number")
		number, exist := numberValue.IntValue()
		if !exist {
			return ginstarter.RespRestBadParameters("number required"), nil
		}
		pager := Pager[D]{
			Number: int(number),
			Size:   int(size),
		}
		condition := paramJson.Get("condition")
		rawConditionJson := condition.RawJsonString()
		param := make(map[string]any)
		flag := b.SetAuthorityLimitMap(request, param)
		if !flag {
			return ginstarter.RespRestUnAuthorized(), nil
		}
		if rawConditionJson != "" {
			if err != nil {
				return nil, err
			}
			json.ParseJsonPanic(rawConditionJson, &param)
			if len(param) == 0 {
				return ginstarter.RespRestBadParameters(), nil
			}
			if !b.checkField(param, query) {
				return ginstarter.RespRestBadParameters(), nil
			}
			param = coll.MapCollect(param, func(k string, v any) (string, any) {
				return str.CamelToSnake(k), v
			})
		}
		err = b.baseBizService.BaseQueryByPager(param, &pager)
		if err != nil {
			return nil, err
		}
		return ginstarter.RespRestSuccess(pager), nil
	}
}

func (b *BaseRouter[ID, S, M, Q, D]) ModifyById() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		id, err := CovertStringToID[ID](request.GetPathParam("id"))
		if err != nil {
			return nil, err
		}

		var update map[string]any
		rawBytes, err := request.GetRawBodyData()
		if err != nil {
			return nil, err
		}
		json.ParseBytesPanic(rawBytes, &update)
		if len(update) == 0 {
			return ginstarter.RespRestBadParameters(), nil
		}
		if !b.checkField(update, modify) {
			return ginstarter.RespRestBadParameters(), nil
		}
		param := map[string]any{"id": id}
		flag := b.SetAuthorityLimitMap(request, param)
		if !flag {
			return ginstarter.RespRestUnAuthorized(), nil
		}
		_, err = b.baseBizService.BaseModifyByID(update, param)
		if err != nil {
			return nil, err
		}
		return ginstarter.RespRestSuccess(), nil
	}
}

func (b *BaseRouter[ID, S, M, Q, D]) RemoveById() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		id, err := CovertStringToID[ID](request.GetPathParam("id"))
		if err != nil {
			return nil, err
		}
		param := map[string]any{"id": id}
		flag := b.SetAuthorityLimitMap(request, param)
		if err != nil {
			return nil, err
		}
		if !flag {
			return ginstarter.RespRestUnAuthorized(), nil
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

// GetAuthorityData 获取当前请求的认证信息
func (s *SimpleRouter[ID]) GetAuthorityData(request *ginstarter.Request, notRequired ...bool) Authority[ID] {
	if s.authorityFetch == nil {
		logger.Logrus().Warningln("not set authority fetch method")
		return nil
	}
	result := s.authorityFetch(request)
	if len(notRequired) > 0 && notRequired[0] {
		return result
	}
	if result == nil {
		request.Panic(ginstarter.StatusCodeUnauthorized, errors.New("Unauthorized Request"))
	}
	return result
}
