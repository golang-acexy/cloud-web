package webcloud

import (
	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/acexy/golang-toolkit/util/reflect"
	"github.com/acexy/golang-toolkit/util/str"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

type BaseRouter[T any, ID IDType] struct {
	baseBizService BaseBizService[T, ID]

	// 安全设置
	updateAllowedFields []string // 允许自由更新的字段
	queryAllowedFields  []string // 允许自由查询的字段
}

// NewBaseRouter 创建基础路由
// notAllowedUpdateFieldNames: 设置不允许更新的字段 (驼峰结构体中的字段名)
func NewBaseRouter[T any, ID IDType](baseBizService BaseBizService[T, ID], notAllowedUpdateFieldNames ...string) BaseRouter[T, ID] {
	var t T
	allFieldNames, err := reflect.AllFieldName(t)
	if err != nil {
		panic(err)
	}
	allFieldNames = coll.SliceCollect(allFieldNames, func(field string) string {
		if field == "ID" || field == "Id" {
			return "id"
		}
		return str.CamelToSnake(str.LowFirstChar(field))
	})
	if len(notAllowedUpdateFieldNames) == 0 {
		notAllowedUpdateFieldNames = []string{
			"id",
			"createdAt",
			"modifyAt",
			"createTime",
			"updateTime",
			"updateAt",
		}
	}
	forbidUpdateNames := coll.SliceCollect(notAllowedUpdateFieldNames, func(field string) string {
		if field == "ID" || field == "Id" {
			return "id"
		}
		return str.CamelToSnake(str.LowFirstChar(field))
	})

	return BaseRouter[T, ID]{
		baseBizService: baseBizService,
		updateAllowedFields: coll.SliceFilter(allFieldNames, func(field string) bool {
			return !coll.SliceContains(forbidUpdateNames, field)
		}),
		queryAllowedFields: allFieldNames,
	}
}

// SetAllowedUpdateFieldNames 设置允许更新的字段 (驼峰结构体中的字段名) 将完全覆盖当前设置
func (b *BaseRouter[T, ID]) SetAllowedUpdateFieldNames(allowedUpdateFieldNames ...string) {
	if len(allowedUpdateFieldNames) > 0 {
		b.updateAllowedFields = coll.SliceCollect(allowedUpdateFieldNames, func(field string) string {
			return str.CamelToSnake(str.LowFirstChar(field))
		})
	}
}

// checkField 安全检查
func (b *BaseRouter[T, ID]) checkField(param map[string]any, isWrite bool) bool {
	allowedFieldNames := func(write bool) []string {
		if isWrite {
			return b.updateAllowedFields
		}
		return b.queryAllowedFields
	}(isWrite)
	input := coll.MapFilterToSlice(param, func(k string, v any) (string, bool) {
		return str.CamelToSnake(k), true
	})
	if !coll.SliceIsSubset(input, allowedFieldNames) {
		logger.Logrus().Warningln("request field not allowed: ", input)
		return false
	}
	return true
}

// RegisterBaseHandler 注册基础路由
func (b *BaseRouter[T, ID]) RegisterBaseHandler(router *ginstarter.RouterWrapper, baseRouter BaseRouter[T, ID]) {
	router.POST1("save", []string{gin.MIMEJSON, gin.MIMEPOSTForm}, baseRouter.save())

	router.GET("by-id/:id", baseRouter.queryById())
	router.POST1("query-one", []string{gin.MIMEJSON, gin.MIMEPOSTForm}, baseRouter.queryOne())
	router.POST1("query-by-page", []string{gin.MIMEJSON, gin.MIMEPOSTForm}, baseRouter.queryById())
	router.PUT1("by-id/:id", []string{gin.MIMEJSON}, baseRouter.updateById())
	router.DELETE("by-id/:id", baseRouter.deleteById())
}

// 基础CRUD

func (b *BaseRouter[T, ID]) save() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		var t T
		request.MustBindBodyJson(&t)
		id, err := b.baseBizService.Save(&t)
		if err != nil {
			logger.Logrus().Errorln("cant save:", t, err)
			return nil, err
		}
		return ginstarter.RespRestSuccess(id), nil
	}
}

func (b *BaseRouter[T, ID]) queryById() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		id, err := CovertStringToID[ID](request.GetPathParam("id"))
		if err != nil {
			return nil, err
		}
		var t T
		row, err := b.baseBizService.QueryByID(id, &t)
		if err != nil {
			return nil, err
		}
		if row > 0 {
			return ginstarter.RespRestSuccess(t), nil
		}
		return ginstarter.RespRestSuccess(), nil
	}
}

func (b *BaseRouter[T, ID]) queryOne() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		var param map[string]any
		rawBytes, err := request.GetRawBodyData()
		if err != nil {
			return nil, err
		}
		json.ParseBytesPanic(rawBytes, &param)
		if len(param) == 0 {
			return ginstarter.RespRestBadParameters(), nil
		}
		if !b.checkField(param, false) {
			return ginstarter.RespRestBadParameters(), nil
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

func (b *BaseRouter[T, ID]) queryByPage() ginstarter.HandlerWrapper {
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
		var conditionMap map[string]any
		if rawConditionJson != "" {
			if err != nil {
				return nil, err
			}
			json.ParseJsonPanic(rawConditionJson, &conditionMap)
			if len(conditionMap) == 0 {
				return ginstarter.RespRestBadParameters(), nil
			}
			if !b.checkField(conditionMap, false) {
				return ginstarter.RespRestBadParameters(), nil
			}
		}
		b.baseBizService.QueryByPager(conditionMap, &pager)
		return ginstarter.RespRestSuccess(pager), nil
	}
}

func (b *BaseRouter[T, ID]) updateById() ginstarter.HandlerWrapper {
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
		if !b.checkField(update, true) {
			return ginstarter.RespRestBadParameters(), nil
		}
		_, err = b.baseBizService.ModifyByID(id, update)
		if err != nil {
			return nil, err
		}
		return ginstarter.RespRestSuccess(), nil
	}
}

func (b *BaseRouter[T, ID]) deleteById() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		id, err := CovertStringToID[ID](request.GetPathParam("id"))
		if err != nil {
			return nil, err
		}
		_, err = b.baseBizService.RemoveByID(id)
		if err != nil {
			return nil, err
		}
		return ginstarter.RespRestSuccess(), nil
	}
}
