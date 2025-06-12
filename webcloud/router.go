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
	baseBizService      BaseBizService[T, ID]
	updateAllowedFields []string // 安全设置
}

func NewBaseRouter[T any, ID IDType](baseBizService BaseBizService[T, ID]) BaseRouter[T, ID] {
	var t T
	field, err := reflect.AllFieldName(t)
	if err != nil {
		panic(err)
	}
	return BaseRouter[T, ID]{
		baseBizService: baseBizService,
		updateAllowedFields: coll.SliceCollect(field, func(field string) string {
			if field == "ID" || field == "Id" {
				return "id"
			}
			return str.CamelToSnake(str.LowFirstChar(field))
		}),
	}
}

// RegisterBaseHandler 注册基础路由
func (b *BaseRouter[T, ID]) RegisterBaseHandler(router *ginstarter.RouterWrapper, baseRouter BaseRouter[T, ID]) {
	router.POST1("save", []string{gin.MIMEJSON, gin.MIMEPOSTForm}, baseRouter.save())

	router.GET("by-id/:id", baseRouter.queryById())
	router.GET("query-one", baseRouter.queryOne())
	router.POST1("query-by-page", []string{gin.MIMEJSON, gin.MIMEPOSTForm}, baseRouter.queryById())
	router.PUT1("by-id/:id", []string{gin.MIMEJSON}, baseRouter.updateById())
	router.DELETE("by-id/:id", baseRouter.deleteById())
}

// checkField 安全检查
func (b *BaseRouter[T, ID]) checkField(param map[string]any) bool {
	input := coll.MapFilterToSlice(param, func(k string, v any) (string, bool) {
		return str.CamelToSnake(k), true
	})
	if !coll.SliceIsSubset(input, b.updateAllowedFields) {
		logger.Logrus().Warningln("request field not allowed: ", input)
		return false
	}
	return true
}

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
		if !b.checkField(param) {
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
			if !b.checkField(conditionMap) {
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
		if !b.checkField(update) {
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
