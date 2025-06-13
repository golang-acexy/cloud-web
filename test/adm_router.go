package test

import (
	"github.com/golang-acexy/cloud-web/webcloud"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

type AdmUserRouter[ID webcloud.IDType, S, M, Q, T any] struct {
	*webcloud.BaseRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]
	bizService webcloud.BaseBizService[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]
}

func NewAdmUserRouter() *AdmUserRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO] {
	var bizService = UserBizService[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]{}
	return &AdmUserRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]{
		BaseRouter: webcloud.NewBaseRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO](bizService),
		bizService: bizService,
	}
}

func (r *AdmUserRouter[ID, S, M, Q, T]) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "adm/user/",
	}
}

func (r *AdmUserRouter[ID, S, M, Q, T]) registerBaseHandler(router *ginstarter.RouterWrapper) {
	r.BaseRouter.RegisterBaseHandler(router, r.BaseRouter)
}

func (r *AdmUserRouter[ID, S, M, Q, T]) Handlers(router *ginstarter.RouterWrapper) {
	// 注册基础路由
	r.registerBaseHandler(router)

	// 自定义实现业务
	router.GET("test", r.test())
}

// 自定义实现业务

func (r *AdmUserRouter[ID, S, M, Q, T]) test() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		//authority := r.GetAuthorityData(request)
		return ginstarter.RespRestSuccess(), nil
	}
}

// 重写基础服务的save方法
func (*AdmUserRouter[ID, S, M, Q, T]) save() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.RespRestSuccess(), nil
	}
}
