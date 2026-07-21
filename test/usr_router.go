package test

import (
	"github.com/golang-acexy/cloud-web/webcloud"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

var usrAuthorityFetch webcloud.AuthorityFetch[uint64] = func(request *ginstarter.Request) webcloud.Authority[uint64] {
	return AuthorityUser[uint64]{id: 12345}
}

type UsrUserRouter struct {
	*webcloud.BaseRouter[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO]
}

func NewUsrUserRouter() *UsrUserRouter {
	return &UsrUserRouter{
		BaseRouter: webcloud.NewBaseRouterWithAuthority[uint64, UserSDTO, UserMDTO, UserQDTO, UserDTO](
			userBizService,
			usrAuthorityFetch,
			webcloud.AuthorityDataField{StructField: "UserID", Column: "user_id"},
		),
	}
}

func (r *UsrUserRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{GroupPath: "/usr/user"}
}

func (r *UsrUserRouter) Handlers(router *ginstarter.RouterWrapper) {
	r.RegisterBaseHandlers(router, r)
}

func (r *UsrUserRouter) QueryOne() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.RespRestSuccess("overridden-query-one"), nil
	}
}
