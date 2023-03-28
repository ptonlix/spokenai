package main

import (
	"context"
	"fmt"

	"github.com/ptonlix/spokenai/configs"
	"github.com/ptonlix/spokenai/internal/engine/console"
	"github.com/ptonlix/spokenai/pkg/env"
	"github.com/ptonlix/spokenai/pkg/logger"
	"github.com/ptonlix/spokenai/pkg/shutdown"
	"github.com/ptonlix/spokenai/pkg/timeutil"
	"go.uber.org/zap"
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
	s, err := console.NewConsoleServer(accessLogger, context.Background())
	if err != nil {
		accessLogger.Error("New ConsoleServer error: ", zap.String("error", fmt.Sprintf("%+v", err)))
		panic(err)
	}
	s.ListenAndServe()
	// 优雅关闭
	shutdown.NewHook().Close(

		func() {
			s.Close()
		},
	)
}
