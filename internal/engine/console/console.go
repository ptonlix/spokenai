package console

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/ptonlix/spokenai/configs"
	"github.com/ptonlix/spokenai/internal/engine/console/io"
	"github.com/ptonlix/spokenai/internal/pkg/logo"
	"github.com/ptonlix/spokenai/internal/pkg/nettest"
	"github.com/ptonlix/spokenai/internal/pkg/rocket"
	"github.com/ptonlix/spokenai/internal/pkg/role"
	hook "github.com/robotn/gohook"
	"go.uber.org/zap"
)

type Resource struct {
	Username      string
	ChoiceTeacher int
}

type Server struct {
	La rocket.Launcher
}

func InputBaseInfo(logger *zap.Logger) *Resource {
	console := io.NewIoer()
	// 输入用户名和选择老师
	resource := Resource{}
	console.Println("--------------------------------Part1--------------------------------")
	console.SlowPrint("请输入您的英文名/Please enter your English name:")
	// fmt.Scanln(&resoure.Username)
	console.GetInput(&resource.Username)
	console.Println("---------------------------------------------------------------------")
	console.Println("老师列表/Teacher List:")
	console.Println("---------------------------------------------------------------------")
	rolemap := role.Get()
	for k, v := range *rolemap {
		console.Println("|" + strconv.Itoa(k) + "|" + fmt.Sprintf("%-65s", "      "+v) + "|")
		console.Println("---------------------------------------------------------------------")
	}
	console.SlowPrint("请选择你对话的英语老师/Please select Number:")
	console.GetInput(&resource.ChoiceTeacher)
	for {
		if _, ok := (*rolemap)[resource.ChoiceTeacher]; !ok {
			console.PrintlnRed("错误: 输入错误,请重新输入/Input Error")
			console.SlowPrint("请选择你对话的英语老师/Please select Number:")
			console.GetInput(&resource.ChoiceTeacher)
		} else {
			break
		}
	}
	logger.Info("Input UserInfo: ", zap.String("username", fmt.Sprintf("%+v", resource.Username)), zap.String("username", fmt.Sprintf("%+v", resource.ChoiceTeacher)))
	return &resource
}

func NewConsoleServer(logger *zap.Logger) (*Server, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	// 展示Logo
	logo.PrintLogo(os.Stdout)

	// 展示说明
	logo.PrintIntroduce(os.Stdout)

	logo.PrintQrcode()

	// 检测网络
	console := io.NewIoer()
	if !nettest.GetResult(logger) {
		console.PrintlnRed("错误： 网络连通性检测失败，请检查日志")
		console.PrintlnRed("ERROR: Network connectivity detection failed, check the logs")
	}

	// 输入用户名和选择老师
	res := InputBaseInfo(logger)
	// 配置
	sysConfig := configs.Get()
	r, err := rocket.New(logger,
		rocket.WithApiKey(sysConfig.OpenAi.Base.ApiKey),
		rocket.WithApiHost(sysConfig.OpenAi.Base.ApiHost),
		rocket.WithAudioModel(sysConfig.OpenAi.Audio.AudioModel),
		rocket.WithChatMaxToken(sysConfig.OpenAi.Chat.ChatMaxToken),
		rocket.WithChatModel(sysConfig.OpenAi.Chat.ChatModel),
		rocket.WithChatTemperature(sysConfig.OpenAi.Chat.ChatTemperature),
		rocket.WithChatTopP(sysConfig.OpenAi.Chat.ChatTopP),
		rocket.WithEnableConsole(),
		rocket.WithDataDir(sysConfig.File.History.Path),
		rocket.WithUserId(res.Username),
		rocket.WithRoleId(strconv.Itoa(res.ChoiceTeacher)),
	)
	if err != nil {
		return nil, err
	}
	return &Server{La: r}, nil
}

func add() {
	fmt.Println("--- Please press ctrl + shift + q to stop hook ---")
	hook.Register(hook.KeyDown, []string{"q", "ctrl", "shift"}, func(e hook.Event) {
		fmt.Println("ctrl-shift-q")
		hook.End()
	})

	fmt.Println("--- Please press w---")
	hook.Register(hook.KeyDown, []string{"w"}, func(e hook.Event) {
		fmt.Println("w")
	})

	fmt.Println("--- Please press space down---")
	hook.Register(hook.KeyDown, []string{"space"}, func(e hook.Event) {
		fmt.Println("space down")
	})

	fmt.Println("--- Please press space up---")
	hook.Register(hook.KeyHold, []string{"space"}, func(e hook.Event) {
		fmt.Println("space up")
	})

	s := hook.Start()
	<-hook.Process(s)
}

func (s *Server) ListenAndServe() error {
	// 唤醒语音输入，转换文字，输出
	//praudio.RecordAndSaveWithInterruptShow("output.wav")
	add()
	return nil
}
