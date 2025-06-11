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
	bizService          BaseBizService[T, ID]
	updateAllowedFields []string // 安全设置
}

func NewBaseRouter[T any, ID IDType](bizService BaseBizService[T, ID]) BaseRouter[T, ID] {
	var t T
	field, err := reflect.AllFieldName(t)
	if err != nil {
		panic(err)
	}
	return BaseRouter[T, ID]{
		bizService: bizService,
		updateAllowedFields: coll.SliceCollect(field, func(field string) string {
			if field == "ID" {
				return "id"
			}
			return str.CamelToSnake(str.LowFirstChar(field))
		}),
	}
}

func (b *BaseRouter[T, ID]) RegisterBaseHandler(router *ginstarter.RouterWrapper, baseRouter BaseRouter[T, ID]) {
	router.POST1("save", []string{gin.MIMEJSON, gin.MIMEPOSTForm}, baseRouter.save())
	router.GET("by-id/:id", baseRouter.queryById())
	router.PUT1("by-id/:id", []string{gin.MIMEJSON}, baseRouter.updateById())
}

func (b *BaseRouter[T, ID]) save() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		var t T
		request.MustBindBodyJson(&t)
		id, err := b.bizService.Save(&t)
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
		row, err := b.bizService.QueryById(id, &t)
		if err != nil {
			return nil, err
		}
		if row > 0 {
			return ginstarter.RespRestSuccess(t), nil
		}
		return ginstarter.RespRestSuccess(), nil
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
		input := coll.MapFilterToSlice(update, func(k string, v any) (string, bool) {
			return str.CamelToSnake(k), true
		})
		if !coll.SliceIsSubset(input, b.updateAllowedFields) {
			logger.Logrus().Warningln("request field not allowed: ", input)
			return ginstarter.RespRestBadParameters(), nil
		}
		_, err = b.bizService.ModifyById(id, update)
		if err != nil {
			return nil, err
		}
		return ginstarter.RespRestSuccess(), nil
	}
}
