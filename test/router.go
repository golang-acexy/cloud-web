package test

import (
	"github.com/golang-acexy/cloud-web/webcloud"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

type User struct {
	ID        uint64 `json:"id"`
	ClassName string `json:"class_name"`
}

type UserRouter[T User, ID uint64] struct {
	webcloud.BaseRouter[User, uint64]
	bizService webcloud.BaseBizService[User, uint64]
}

func NewUserRouter() *UserRouter[User, uint64] {
	var bizService = UserService[User, uint64]{}
	return &UserRouter[User, uint64]{
		BaseRouter: webcloud.NewBaseRouter[User, uint64](bizService),
		bizService: bizService,
	}
}

func (u UserRouter[User, uint64]) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "user",
	}
}

func (u UserRouter[User, uint64]) registerBaseHandler(router *ginstarter.RouterWrapper) {
	u.BaseRouter.RegisterBaseHandler(router, u.BaseRouter)
}

func (u UserRouter[User, uint64]) Handlers(router *ginstarter.RouterWrapper) {
	// 注册基础路由
	u.registerBaseHandler(router)

	// 自定义实现业务
	router.GET("test", u.test())
}

// 自定义实现业务

func (UserRouter[User, uint64]) test() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.RespRestSuccess(), nil
	}
}
