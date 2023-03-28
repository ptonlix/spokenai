package console

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"unicode"

	"github.com/ptonlix/spokenai/configs"
	"github.com/ptonlix/spokenai/internal/engine/console/io"
	"github.com/ptonlix/spokenai/internal/pkg/clean"
	"github.com/ptonlix/spokenai/internal/pkg/logo"
	"github.com/ptonlix/spokenai/internal/pkg/manaudio"
	"github.com/ptonlix/spokenai/internal/pkg/nettest"
	"github.com/ptonlix/spokenai/internal/pkg/rocket"
	"github.com/ptonlix/spokenai/internal/pkg/role"
	hook "github.com/robotn/gohook"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	instructions = "Warning:\n1.在终端使用存在易用性问题，请严格遵守提示信息来输入,\n2.不能随意敲击键盘,建议关闭输入法,避免出现不可用现象 \n3.请在使用前,检查配置文件是否配置在正确,如密钥\n4.如有疑问, 请联系作者, 谢谢~\n\n"
	recordaudio  = "\r---> 请按W键开始获取音频输入 按Q键可以停止音频输入,最长录制60s 按ESC返回上级菜单"
	start        = "\r---> 请输入S键并回车, 开始输入您的英文名并选择您喜欢的AI老师 按ESC退出SpokenAI"
)

const (
	recordMenu = -1 //正在录音可以中断
	startMenu  = 0
	speakMenu  = 1
	chatMenu   = 2
	endMenu    = 3
)

type Resource struct {
	Username      string
	ChoiceTeacher int
}

type Server struct {
	La            rocket.Launcher
	Audio         *manaudio.Manager
	InputContent  string
	OutputContent string
	UseInfo       *Resource
	console       io.Ioer
	logger        *zap.Logger
	sig           chan struct{}
	menuId        int
	ctx           context.Context
	cancelRecord  context.CancelFunc
	cleantool     clean.Cleaner
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

func NewConsoleServer(logger *zap.Logger, ctx context.Context) (*Server, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	// 展示Logo
	logo.PrintLogo(os.Stdout)

	// 展示说明
	logo.PrintIntroduce(os.Stdout)

	// 检测网络
	console := io.NewIoer()
	go console.SlowPrint("\nTesting network.................")
	if !nettest.GetResult(logger) {
		console.PrintlnRed("\r错误: 网络连通性检测失败，请检查日志")
		return nil, errors.New("network link failed")
	} else {
		console.PrintlnBlue("\rGood: Network connectivity detection is ok\n")
	}
	// 配置
	sysConfig := configs.Get()
	tool, err := clean.NewCleaner(logger,
		clean.WithDataChatDir(sysConfig.File.History.Path),
		clean.WithDataPlayDir(sysConfig.File.Audio.Play.Path),
		clean.WithDataRecordDir(sysConfig.File.Audio.Record.Path),
	)
	if err != nil {
		console.PrintlnRed("\r错误: 初始化服务失败,请检查日志")
		return nil, errors.New("init cleaner failed")
	}
	return &Server{console: io.NewIoer(), logger: logger, sig: make(chan struct{}), ctx: ctx, menuId: 0, cleantool: tool}, nil
}

func (s *Server) CleanData() {
	s.console.PrintlnYellow("\r---> 如需要删除音频和对话数据，请输入[y/n]")
	str := ""
	s.console.GetInput(&str)
	for {
		if str == "y" {
			if err := s.cleantool.ClearAllData(); err != nil {
				s.logger.Error("Clear AllData Failed", zap.String("error", fmt.Sprintf("%+v", err)))
				s.console.PrintlnRed("\r---> 删除数据失败,请手动删除data目录下的数据")
			}
			os.Exit(0)
		} else if str == "n" {
			os.Exit(0)
		} else {
			s.console.Print("\r---> 输入错误,请重新输入")
			s.console.GetInput(&str)
		}
	}
}

func (s *Server) Close() error {
	if s.cancelRecord != nil {
		s.cancelRecord()
	}
	if s.La != nil {
		if err := s.La.Close(); err != nil {
			return err
		} //备份聊天纪录
	}
	if s.Audio != nil {
		if err := s.Audio.Close(); err != nil {
			return err
		} //备份聊天纪录
	}
	return nil
}

func (s *Server) keyboardEvent() bool {
	flag := false
	// 开始输入个人信息和选择老师
	hook.Register(hook.KeyDown, []string{"s"}, func(e hook.Event) {
		s.console.Flush()
		if s.menuId == startMenu {
			// 输入用户名和选择老师
			res := s.inputBaseInfo()
			// 配置
			sysConfig := configs.Get()
			audioManager, _ := manaudio.NewManager(s.logger,
				manaudio.WithRoleId(strconv.Itoa(res.ChoiceTeacher)),
				manaudio.WithUserId(res.Username),
				manaudio.WithDataRecordDir(sysConfig.File.Audio.Record.Path),
				manaudio.WithDataPlayDir(sysConfig.File.Audio.Play.Path),
				manaudio.WithAudioHost(sysConfig.File.Audio.Play.TtsHost),
				manaudio.WithEnablePlay(sysConfig.File.Audio.Play.Enable),
			)

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
				s.console.PrintlnRed("\r---> ERROR: 系统错误,请重新输入S键并回车, 开始输入您的英文名并选择您喜欢的AI老师")
			} else {
				s.La = rocketServer
				s.Audio = audioManager
				s.UseInfo = res
				s.menuId = speakMenu //跳到下个菜单
				s.console.PrintlnBlue(recordaudio)
			}
			hook.End()
		}
	})
	// 返回上级菜单
	hook.Register(hook.KeyDown, []string{"esc"}, func(e hook.Event) {
		if s.menuId == startMenu {
			s.console.PrintlnRed("\r---> SpokenAI之旅已结束,请按Ctrl+C关闭程序,希望您再次使用")
			s.console.PrintlnYellow("\r---> 如需要删除音频和对话数据,请按Y或者N键")
			s.menuId = endMenu
		} else if s.menuId == speakMenu {
			go s.La.Close()    //备份聊天纪录
			go s.Audio.Close() //备份音频记录
			s.menuId = startMenu
			s.console.PrintlnBlue(start)
			hook.End()
		}
	})
	// 开始口语练习
	hook.Register(hook.KeyDown, []string{"w"}, func(e hook.Event) {
		if s.menuId == speakMenu {
			go func() {
				s.menuId = recordMenu
				ctx, cancel := context.WithCancel(s.ctx)
				s.cancelRecord = cancel
				s.Audio.RecordAudioWithContext(ctx, s.sig)
				// 发送请求获取音频转文W
				go s.console.SlowPrint("\rLoading............")
				var err error
				s.InputContent, err = s.La.Translate(s.Audio.GetRecordAudio())
				if err != nil {
					s.logger.Error("Translate Request error:", zap.String("error", fmt.Sprintf("%+v", err)))
					s.console.PrintlnRed("\r---> ERROR: 请求音频翻译出错, 请按W键开始获取音频输入 按Q键可以停止音频输入,最长录制60s")
				} else {
					s.console.Println(fmt.Sprintf("\r%s : %s", s.UseInfo.Username, s.InputContent))
					s.console.PrintlnBlue("\r---> 如需修改请按W重新录音, 无需修改请按ctrl+shift+enter发送")
				}
				s.menuId = speakMenu
			}()
			hook.End()

		}
	})
	// 中断语音输入
	hook.Register(hook.KeyDown, []string{"q"}, func(e hook.Event) {
		if s.menuId == recordMenu {
			if s.cancelRecord != nil {
				s.cancelRecord()
				s.cancelRecord = nil
			}
			s.menuId = speakMenu
			hook.End()
		}
	})

	// 对话
	hook.Register(hook.KeyDown, []string{"ctrl", "shift", "enter"}, func(e hook.Event) {
		if s.menuId == speakMenu && len(s.InputContent) != 0 {
			s.menuId = chatMenu
			go s.console.SlowPrint("\rLoading............")
			outputContent, totalToken, err := s.La.Chat(s.InputContent)
			if err != nil {
				s.logger.Error("Chat Request error:", zap.String("error", fmt.Sprintf("%+v", err)))
				s.console.PrintlnRed("\r请求对话出错,请重新按ctrl+shift+enter发送")
			} else {
				s.Audio.CallTTSserver(outputContent)
				go s.Audio.PlayAudio()
				s.console.SlowPrint(fmt.Sprintf("\r%s : %s\n", "Teacher", outputContent))
				warnToken, _ := decimal.NewFromFloat(0.8).Add(decimal.NewFromFloat(float64(configs.Get().OpenAi.Chat.ChatMaxToken))).Float64()
				if float64(totalToken) > warnToken {
					s.console.PrintlnYellow(fmt.Sprintf("Warning: Total Token Count: %d, MaxToken is about to be reached, Please Restart.", totalToken))
				} else {
					s.console.PrintlnBlue(fmt.Sprintf("\r---> Total Token Count: %d", totalToken))
				}
				s.console.PrintlnBlue(recordaudio)
			}
			s.menuId = speakMenu
			hook.End()
		}
	})

	hook.Register(hook.KeyDown, []string{"t"}, func(e hook.Event) {
		s.console.PrintlnBlue(fmt.Sprintf("MenuId = %d", s.menuId))
		hook.End()
	})

	hook.Register(hook.KeyDown, []string{"y"}, func(e hook.Event) {
		if s.menuId == endMenu {
			if err := s.cleantool.ClearAllData(); err != nil {
				s.logger.Error("Clear AllData Failed", zap.String("error", fmt.Sprintf("%+v", err)))
				s.console.PrintlnRed("\r---> 删除数据失败,请手动删除data目录下的数据")
			} else {
				s.console.PrintlnBlue("\r---> 删除数据成功")
			}
			s.console.PrintlnRed("\r---> SpokenAI之旅已结束,请按Ctrl+C关闭程序,希望您再次使用")
			flag = true
			hook.End()
		}
	})
	hook.Register(hook.KeyDown, []string{"n"}, func(e hook.Event) {
		if s.menuId == endMenu {
			s.console.PrintlnRed("\r---> SpokenAI之旅已结束,请按Ctrl+C关闭程序,希望您再次使用")
			flag = true
			hook.End()
		}
	})

	e := hook.Start()
	<-hook.Process(e)
	return flag
}

func (s *Server) ListenAndServe() error {
	// 唤醒语音输入，转换文字，输出
	s.console.PrintlnYellow(instructions)
	s.console.SlowPrint("---> 让开始我们的SpokenAI之旅吧~\n")
	s.console.PrintlnBlue(start)

	for !s.keyboardEvent() {
	}

	return nil
}
