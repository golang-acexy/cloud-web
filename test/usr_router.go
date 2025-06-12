package test

import (
	"github.com/golang-acexy/cloud-web/webcloud"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

var usrAuthorityFetch webcloud.AuthorityFetch[uint64] = func(request *ginstarter.Request) webcloud.Authority[uint64] {
	return AuthorityUser[uint64]{
		id: 12345,
	}
}

type UsrUserRouter[ID webcloud.IDType, S, M, Q, T any] struct {
	*webcloud.BaseRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]
	bizService webcloud.BaseBizService[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]
}

func NewUsrUserRouter() *UsrUserRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO] {
	var bizService = UserBizService[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]{}
	return &UsrUserRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]{
		BaseRouter: webcloud.NewBaseRouterWithAuthority[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO](bizService, usrAuthorityFetch, "user_id"),
		bizService: bizService,
	}
}

func (u *UsrUserRouter[ID, S, M, Q, T]) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "usr/user",
	}
}

func (u *UsrUserRouter[ID, S, M, Q, T]) registerBaseHandler(router *ginstarter.RouterWrapper) {
	u.BaseRouter.RegisterBaseHandler(router, u.BaseRouter)
}

func (u *UsrUserRouter[ID, S, M, Q, T]) Handlers(router *ginstarter.RouterWrapper) {
	// 注册基础路由
	u.registerBaseHandler(router)

	// 自定义实现业务
	router.GET("test", u.test())
}

// 自定义实现业务

func (*UsrUserRouter[ID, S, M, Q, T]) test() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.RespRestSuccess(), nil
	}
}
