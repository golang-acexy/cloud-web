package test

import (
	"github.com/golang-acexy/cloud-web/webcloud"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

type UserRouter[ID webcloud.IDType, S, M, Q, T any] struct {
	webcloud.BaseRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]
	bizService webcloud.BaseBizService[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]
}

func NewUserRouter() *UserRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO] {
	var bizService = UserBizService[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]{}
	return &UserRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]{
		BaseRouter: webcloud.NewBaseRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO](bizService),
		bizService: bizService,
	}
}

func (u *UserRouter[ID, S, M, Q, T]) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "user",
	}
}

func (u *UserRouter[ID, S, M, Q, T]) registerBaseHandler(router *ginstarter.RouterWrapper) {
	u.BaseRouter.RegisterBaseHandler(router, u.BaseRouter)
}

func (u *UserRouter[ID, S, M, Q, T]) Handlers(router *ginstarter.RouterWrapper) {
	// 注册基础路由
	u.registerBaseHandler(router)

	// 自定义实现业务
	router.GET("test", u.test())
}

// 自定义实现业务

func (*UserRouter[ID, S, M, Q, T]) test() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.RespRestSuccess(), nil
	}
}

// 重写基础服务的save方法
func (*UserRouter[ID, S, M, Q, T]) save() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.RespRestSuccess(), nil
	}
}
