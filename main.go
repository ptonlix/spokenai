package main

import (
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
	s, _ := console.NewConsoleServer(accessLogger)
	s.ListenAndServe()
	// 优雅关闭
	shutdown.NewHook().Close(

		func() {
			if s.La != nil {
				if err := s.La.Close(); err != nil {
					accessLogger.Error("La close err", zap.Error(err))
				}
			}
		},
	)
}
