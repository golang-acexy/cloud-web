package test

import (
	"github.com/golang-acexy/starter-gin/ginstarter"
	"github.com/golang-acexy/starter-parent/parent"
)

var starterLoader *parent.StarterLoader

func init() {
	starterLoader = parent.InitStarterLoader([]parent.Starter{
		&ginstarter.GinStarter{
			Config: ginstarter.GinConfig{
				ListenAddress: "127.0.0.1:0",
				Routers: []ginstarter.Router{
					NewUsrUserRouter(),
					NewAdmUserRouter(),
				},
			},
		},
	})
}
