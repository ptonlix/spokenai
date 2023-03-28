package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ptonlix/spokenai/configs"
	"github.com/ptonlix/spokenai/internal/engine/console"
	"github.com/ptonlix/spokenai/internal/pkg/clean"
	"github.com/ptonlix/spokenai/internal/pkg/logo"
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

	ClearData(accessLogger)

	ShowQr()

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

func ClearData(logger *zap.Logger) {
	if !env.ClearDataFlag() {
		return
	}
	sysConfig := configs.Get()
	tool, err := clean.NewCleaner(logger,
		clean.WithDataChatDir(sysConfig.File.History.Path),
		clean.WithDataPlayDir(sysConfig.File.Audio.Play.Path),
		clean.WithDataRecordDir(sysConfig.File.Audio.Record.Path),
	)
	if err != nil {
		logger.Error("New Cleaner error: ", zap.String("error", fmt.Sprintf("%+v", err)))
		panic(err)
	}
	if err := tool.ClearAllData(); err != nil {
		logger.Error("Clear AllData Failed", zap.String("error", fmt.Sprintf("%+v", err)))
		panic(err)
	}
	os.Exit(0)
}

func ShowQr() {
	if !env.WxqrShowFlag() {
		return
	}
	logo.PrintQrcode()
	os.Exit(0)
}
