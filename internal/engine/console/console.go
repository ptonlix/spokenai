package console

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"unicode"

	"github.com/ptonlix/spokenai/configs"
	"github.com/ptonlix/spokenai/internal/engine/console/io"
	"github.com/ptonlix/spokenai/internal/pkg/logo"
	"github.com/ptonlix/spokenai/internal/pkg/nettest"
	"github.com/ptonlix/spokenai/internal/pkg/rocket"
	"github.com/ptonlix/spokenai/internal/pkg/role"
	"github.com/ptonlix/spokenai/pkg/praudio"
	hook "github.com/robotn/gohook"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	recordMenu = -1 //正在录音可以中断
	startMenu  = 0
	speakMenu  = 1
	chatMenu   = 2
)

type Resource struct {
	Username      string
	ChoiceTeacher int
}

type Server struct {
	La            rocket.Launcher
	InputContent  string
	OutputContent string
	UseInfo       *Resource
	console       io.Ioer
	logger        *zap.Logger
	sig           chan int
}

func IsChinese(str string) bool {
	var count int
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			count++
			break
		}
	}
	return count > 0
}

func (s *Server) inputModify() string {
	s.console.Flush()
	modifyContent := ""
	s.console.GetInput(&modifyContent)
	fmt.Println(modifyContent)
	for IsChinese(modifyContent) {
		s.console.PrintlnRed("错误: 输入非英文,请重新输入/Input Error")
		s.console.GetInput(&modifyContent)
	}
	return modifyContent
}

func (s *Server) inputBaseInfo() *Resource {
	// 输入用户名和选择老师
	resource := Resource{}
	s.console.Println("--------------------------------Part1--------------------------------")
	s.console.SlowPrint("请输入您的英文名/Please enter your English name:")
	// fmt.Scanln(&resoure.Username)

	s.console.GetInput(&resource.Username)
	s.console.Println("---------------------------------------------------------------------")
	s.console.Println("老师列表/Teacher List:")
	s.console.Println("---------------------------------------------------------------------")
	rolemap := role.Get()
	for k, v := range *rolemap {
		s.console.Println("|" + strconv.Itoa(k) + "|" + fmt.Sprintf("%-65s", "      "+v) + "|")
		s.console.Println("---------------------------------------------------------------------")
	}
	s.console.SlowPrint("请选择你对话的英语老师/Please select Number:")
	s.console.GetInput(&resource.ChoiceTeacher)
	for {
		if _, ok := (*rolemap)[resource.ChoiceTeacher]; !ok {
			s.console.PrintlnRed("错误: 输入错误,请重新输入/Input Error")
			s.console.SlowPrint("请选择你对话的英语老师/Please select Number:")
			s.console.GetInput(&resource.ChoiceTeacher)
		} else {
			break
		}
	}
	s.logger.Info("Input UserInfo: ", zap.String("username", fmt.Sprintf("%+v", resource.Username)), zap.String("username", fmt.Sprintf("%+v", resource.ChoiceTeacher)))
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
		//TODO return
	}
	return &Server{console: io.NewIoer(), logger: logger, sig: make(chan int)}, nil
}

func (s *Server) keyboardEvent(menuId *int) bool {
	flag := false
	// 开始输入个人信息和选择老师
	hook.Register(hook.KeyDown, []string{"s"}, func(e hook.Event) {
		s.console.Flush()
		if *menuId == startMenu {
			// 输入用户名和选择老师
			res := s.inputBaseInfo()
			// 配置
			sysConfig := configs.Get()
			rocketServer, err := rocket.New(s.logger,
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
				s.logger.Error("New RocketServer error:", zap.String("error", fmt.Sprintf("%+v", err)))
				s.console.PrintlnRed("系统错误，请重试～")
			} else {
				s.La = rocketServer
				s.UseInfo = res
				*menuId = speakMenu //跳到下个菜单
				s.console.PrintlnBlue("---> 请按W键开始获取音频输入 按Q键可以停止音频输入,最长录制60s")
			}
			hook.End()
		}
	})
	// 返回上级菜单
	hook.Register(hook.KeyDown, []string{"esc"}, func(e hook.Event) {
		if *menuId == startMenu {
			s.console.PrintlnRed("\r服务已退出,请按Ctrl+C关闭程序")
			flag = true
			hook.End()
		} else if *menuId == 1 {
			s.La.Close() //备份聊天纪录
			*menuId = startMenu
			s.console.PrintlnBlue("\r---> 请按S按键, 输入信息")
			hook.End()
		}
	})
	// 开始口语练习
	hook.Register(hook.KeyDown, []string{"w"}, func(e hook.Event) {
		if *menuId == speakMenu {
			go func() {
				*menuId = recordMenu
				praudio.RecordAndSaveWithShow("output.wav", s.sig)
				// 发送请求获取音频转文W
				go s.console.SlowPrint("\rLoading............")
				var err error
				s.InputContent, err = s.La.Translate("./output.wav")
				if err != nil {
					s.logger.Error("Translate Request error:", zap.String("error", fmt.Sprintf("%+v", err)))
					s.console.PrintlnRed("请求音频翻译出错，请重试～")
				} else {
					s.console.Println(fmt.Sprintf("\r%s : %s", s.UseInfo.Username, s.InputContent))
					s.console.PrintlnBlue("\r如需修改请按W重新录音, 无需修改请按ctrl+shift+enter发送")
				}
				*menuId = speakMenu
			}()
			hook.End()

		}
	})
	// 中断语音输入
	hook.Register(hook.KeyDown, []string{"q"}, func(e hook.Event) {
		if *menuId == recordMenu {
			s.sig <- 1
			*menuId = speakMenu
			hook.End()
		}
	})

	// 对话
	hook.Register(hook.KeyDown, []string{"ctrl", "shift", "enter"}, func(e hook.Event) {
		if *menuId == speakMenu && len(s.InputContent) != 0 {
			*menuId = chatMenu
			go s.console.SlowPrint("\rLoading............")
			outputContent, totalToken, err := s.La.Chat(s.InputContent)
			if err != nil {
				s.logger.Error("Chat Request error:", zap.String("error", fmt.Sprintf("%+v", err)))
				s.console.PrintlnRed("请求对话出错，请重试～")
			} else {
				s.console.SlowPrint(fmt.Sprintf("\r%s : %s\n", "Teacher", outputContent))
				warnToken, _ := decimal.NewFromFloat(0.8).Add(decimal.NewFromFloat(float64(configs.Get().OpenAi.Chat.ChatMaxToken))).Float64()
				if float64(totalToken) > warnToken {
					s.console.PrintlnYellow(fmt.Sprintf("Warning: Total Token Count: %d, MaxToken is about to be reached, Please Restart.", totalToken))
				} else {
					s.console.PrintlnBlue(fmt.Sprintf("Total Token Count: %d", totalToken))
				}
			}
			*menuId = speakMenu
			hook.End()
		}
	})

	e := hook.Start()
	<-hook.Process(e)
	return flag
}

func (s *Server) ListenAndServe() error {
	// 唤醒语音输入，转换文字，输出
	s.console.PrintlnBlue("---> 请按S按键, 输入信息")
	menuId := 0
	for !s.keyboardEvent(&menuId) {
	}

	return nil
}
