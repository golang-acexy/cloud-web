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
	"modify_at",
	"modify_time",
	"update_time",
	"update_at",
}

type BaseRouter[ID IDType, S, M, Q, T any] struct {
	baseBizService BaseBizService[ID, S, M, Q, T]

	// 权限控制
	authorityFetch  AuthorityFetch[ID]
	authorityCheck  bool
	authorityColumn string

	// 字段安全设置
	modifyAllowedColumns []string // 允许自由更新的字段
	queryAllowedColumns  []string // 允许自由查询的字段
	saveAllowedColumns   []string // 允许自由保存的字段
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
func NewBaseRouter[ID IDType, S, M, Q, T any](baseBizService BaseBizService[ID, S, M, Q, T]) *BaseRouter[ID, S, M, Q, T] {
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

	return &BaseRouter[ID, S, M, Q, T]{
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
// column 数据权限控制字段
func NewBaseRouterWithAuthority[ID IDType, S, M, Q, T any](baseBizService BaseBizService[ID, S, M, Q, T], authorityFetch AuthorityFetch[ID], column string) *BaseRouter[ID, S, M, Q, T] {
	router := NewBaseRouter[ID, S, M, Q, T](baseBizService)
	router.authorityFetch = authorityFetch
	router.authorityCheck = true
	router.authorityColumn = column
	if !coll.SliceContains(router.queryAllowedColumns, column) {
		router.queryAllowedColumns = append(router.queryAllowedColumns, column)
	}
	return router
}

func (b *BaseRouter[ID, S, M, Q, T]) convertJsonToMap(request *ginstarter.Request, m mode) (map[string]any, error) {
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
	return param, nil
}

// checkField 安全检查
func (b *BaseRouter[ID, S, M, Q, T]) checkField(param map[string]any, m mode) bool {
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
		logger.Logrus().Warningln("request field not allowed: ", input)
		return false
	}
	return true
}

// 设置数据权限控制字段
func (b *BaseRouter[ID, S, M, Q, T]) setAuthorityLimit(request *ginstarter.Request, param map[string]any) bool {
	if b.authorityCheck {
		authority := b.AuthorityFetch(request)
		if authority == nil {
			return false
		}
		param[b.authorityColumn] = authority.GetIdentityID()
	}
	return true
}

// RegisterBaseHandler 注册基础路由
func (b *BaseRouter[ID, S, M, Q, T]) RegisterBaseHandler(router *ginstarter.RouterWrapper, baseRouter *BaseRouter[ID, S, M, Q, T]) {
	router.POST1("save", []string{gin.MIMEJSON}, baseRouter.save())

	// 通过主键查询单条数据
	router.GET("by-id/:id", baseRouter.queryById())
	// 通过条件查询单条数据
	router.POST1("query-one", []string{gin.MIMEJSON}, baseRouter.queryOne())
	// 通过条件分页查询
	router.POST1("query-by-page", []string{gin.MIMEJSON}, baseRouter.queryById())
	// 通过主键更新数据
	router.PUT1("by-id/:id", []string{gin.MIMEJSON}, baseRouter.updateById())
	// 通过主键删除数据
	router.DELETE("by-id/:id", baseRouter.deleteById())
}

// AuthorityFetch 获取当前请求的认证信息
func (b *BaseRouter[ID, S, M, Q, T]) AuthorityFetch(request *ginstarter.Request) Authority[ID] {
	if b.authorityFetch == nil {
		logger.Logrus().Warningln("no set authority fetch method")
		return nil
	}
	return b.authorityFetch(request)
}

// 基础CRUD

func (b *BaseRouter[ID, S, M, Q, T]) save() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		param, err := b.convertJsonToMap(request, save)
		if err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		flag := b.setAuthorityLimit(request, param)
		if !flag {
			return ginstarter.RespRestUnAuthorized(), nil
		}
		id, err := b.baseBizService.Save(param)
		if err != nil {
			logger.Logrus().Errorln("cant save:", param, err)
			return nil, err
		}
		return ginstarter.RespRestSuccess(id), nil
	}
}

func (b *BaseRouter[ID, S, M, Q, T]) queryById() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		id, err := CovertStringToID[ID](request.GetPathParam("id"))
		if err != nil {
			return nil, err
		}
		param := map[string]any{"id": id}
		flag := b.setAuthorityLimit(request, param)
		if !flag {
			return ginstarter.RespRestUnAuthorized(), nil
		}
		var t T
		row, err := b.baseBizService.QueryByID(param, &t)
		if err != nil {
			return nil, err
		}
		if row > 0 {
			return ginstarter.RespRestSuccess(t), nil
		}
		return ginstarter.RespRestSuccess(), nil
	}
}

func (b *BaseRouter[ID, S, M, Q, T]) queryOne() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		param, err := b.convertJsonToMap(request, query)
		if err != nil {
			return ginstarter.RespRestBadParameters(), nil
		}
		flag := b.setAuthorityLimit(request, param)
		if !flag {
			return ginstarter.RespRestUnAuthorized(), nil
		}
		var t T
		row, err := b.baseBizService.QueryOne(param, &t)
		if err != nil {
			return nil, err
		}
		if row == 0 {
			return ginstarter.RespRestSuccess(), nil
		}
		return ginstarter.RespRestSuccess(t), nil
	}
}

func (b *BaseRouter[ID, S, M, Q, T]) queryByPage() ginstarter.HandlerWrapper {
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
		pager := Pager[T]{
			Number: int(number),
			Size:   int(size),
		}
		condition := paramJson.Get("condition")
		rawConditionJson := condition.RawJsonString()
		var param map[string]any
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
			flag := b.setAuthorityLimit(request, param)
			if !flag {
				return ginstarter.RespRestUnAuthorized(), nil
			}
		}
		b.baseBizService.QueryByPager(param, &pager)
		return ginstarter.RespRestSuccess(pager), nil
	}
}

func (b *BaseRouter[ID, S, M, Q, T]) updateById() ginstarter.HandlerWrapper {
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
		flag := b.setAuthorityLimit(request, param)
		if !flag {
			return ginstarter.RespRestUnAuthorized(), nil
		}

		_, err = b.baseBizService.ModifyByID(param, update)
		if err != nil {
			return nil, err
		}
		return ginstarter.RespRestSuccess(), nil
	}
}

func (b *BaseRouter[ID, S, M, Q, T]) deleteById() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		id, err := CovertStringToID[ID](request.GetPathParam("id"))
		if err != nil {
			return nil, err
		}
		param := map[string]any{"id": id}
		flag := b.setAuthorityLimit(request, param)
		if !flag {
			return ginstarter.RespRestUnAuthorized(), nil
		}
		_, err = b.baseBizService.RemoveByID(param)
		if err != nil {
			return nil, err
		}
		return ginstarter.RespRestSuccess(), nil
	}
}
