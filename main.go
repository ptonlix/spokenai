package main

import (
	"fmt"

	"github.com/ptonlix/spokenai/configs"
	"github.com/ptonlix/spokenai/pkg/env"
	"github.com/ptonlix/spokenai/pkg/logger"
	"github.com/ptonlix/spokenai/pkg/shutdown"
	"github.com/ptonlix/spokenai/pkg/timeutil"
)

func main() {
	// 初始化 access logger
	accessLogger, err := logger.NewJSONLogger(
		logger.WithDisableConsole(),
		logger.WithField("domain", fmt.Sprintf("%s[%s]", configs.ProjectName, env.Active().Value())),
		logger.WithTimeLayout(timeutil.CSTLayout),
		logger.WithFileP(configs.ProjectLogFile),
	)
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	defer func() {
		_ = accessLogger.Sync()
	}()

	// 优雅关闭
	shutdown.NewHook().Close(
	// 关闭
	)
}
