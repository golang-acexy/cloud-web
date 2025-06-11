package webcloud

import "github.com/golang-acexy/starter-gin/ginstarter"

type BaseRouter[T any, ID IDType] struct {
	bizService BaseBizService[T, ID]
}

func NewBaseRouter[T any, ID IDType](bizService BaseBizService[T, ID]) BaseRouter[T, ID] {
	return BaseRouter[T, ID]{
		bizService: bizService,
	}
}

func (b *BaseRouter[T, ID]) RegisterBaseHandler(router *ginstarter.RouterWrapper, baseRouter BaseRouter[T, ID]) {
	router.GET("by-id/:id", baseRouter.queryById())
}

func (b *BaseRouter[T, ID]) queryById() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		id, err := CovertStringToID[ID](request.GetPathParam("id"))
		if err != nil {
			return nil, err
		}
		var t T
		row := b.bizService.QueryById(id, &t)
		if row > 0 {
			return ginstarter.RespRestSuccess(t), nil
		}
		return ginstarter.RespRestSuccess(), nil
	}
}
