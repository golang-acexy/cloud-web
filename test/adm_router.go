package test

import (
	"github.com/golang-acexy/cloud-web/webcloud"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

var userBizService = &UserBizService{}

type AdmUserRouter struct {
	*webcloud.BaseRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]
}

func NewAdmUserRouter() *AdmUserRouter {
	return &AdmUserRouter{
		BaseRouter: webcloud.NewBaseRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO](userBizService),
	}
}

func (r *AdmUserRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{GroupPath: "/adm/user"}
}

func (r *AdmUserRouter) Handlers(router *ginstarter.RouterWrapper) {
	r.RegisterBaseHandlers(router, r)
}
