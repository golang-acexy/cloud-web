package test

import (
	"fmt"
	"github.com/acexy/golang-toolkit/sys"
	"github.com/golang-acexy/starter-gin/ginstarter"
	"github.com/golang-acexy/starter-parent/parent"
	"testing"
)

var starterLoader *parent.StarterLoader

func init() {
	starterLoader = parent.NewStarterLoader([]parent.Starter{
		&ginstarter.GinStarter{
			Config: ginstarter.GinConfig{
				ListenAddress:     ":8080",
				UseReusePortModel: true,
				DebugModule:       true,
				Routers: []ginstarter.Router{
					NewUserRouter(),
				},
				EnableGoroutineTraceIdResponse: true,
			},
		},
	})
	err := starterLoader.Start()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	sys.ShutdownHolding()
}

func TestRun(t *testing.T) {

}
