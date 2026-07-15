package test

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if err := starterLoader.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "启动 Gin Starter 失败: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	if _, err := starterLoader.StopAllBySetting(10 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "停止 Gin Starter 失败: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}
