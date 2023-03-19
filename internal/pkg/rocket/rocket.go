package rocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/ptonlix/spokenai/internal/pkg/role"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type Option func(*option)

type option struct {
	apiKey          string
	apiHost         string
	chatModel       string
	audioModel      string
	chatMaxToken    int
	chatTemperature int
	chatTopP        int

	//Chat Config
	roleId string
	userId string

	//Data Config
	enableConsole bool
	dataDir       string
}

func WithApiKey(apikey string) Option {
	return func(opt *option) {
		opt.apiKey = apikey
	}
}

func WithApiHost(apihost string) Option {
	return func(opt *option) {
		opt.apiHost = apihost
	}
}

func WithChatModel(chatmodel string) Option {
	return func(opt *option) {
		opt.chatModel = chatmodel
	}
}

func WithAudioModel(audiomodel string) Option {
	return func(opt *option) {
		opt.audioModel = audiomodel
	}
}

func WithChatMaxToken(maxtoken int) Option {
	return func(opt *option) {
		opt.chatMaxToken = maxtoken
	}
}

func WithChatTemperature(temperature int) Option {
	return func(opt *option) {
		opt.chatTemperature = temperature
	}
}

func WithChatTopP(topp int) Option {
	return func(opt *option) {
		opt.chatTopP = topp
	}
}

func WithUserId(id string) Option {
	return func(opt *option) {
		opt.userId = id
	}
}

func WithRoleId(id string) Option {
	return func(opt *option) {
		opt.roleId = id
	}
}

func WithEnableConsole() Option {
	return func(opt *option) {
		opt.enableConsole = true
	}
}

func WithDataDir(path string) Option {
	return func(opt *option) {
		opt.dataDir = path
	}
}

type Launcher interface {
	Chat(string) (string, int, error)
	Translate(string) (string, error)
	Moderations(string) (bool, error)
	Close() error
}

func New(logger *zap.Logger, options ...Option) (Launcher, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}
	// 初始化参数
	opt := new(option)
	for _, f := range options {
		f(opt)
	}
	// 初始化OpenAI客户端
	openaiConfig := openai.DefaultConfig(opt.apiKey)
	if len(opt.apiHost) != 0 {
		openaiConfig.BaseURL = opt.apiHost //启用代理
	}
	client := openai.NewClientWithConfig(openaiConfig)

	// 初始化io

	ioclient := NewIO(opt.enableConsole, opt)

	la := &openaiLauncher{opt: opt, client: client, logger: logger, io: ioclient}

	// 初始化聊天数据
	index, _ := strconv.Atoi(opt.roleId)
	if err := la.SaveHistory(openai.ChatMessageRoleSystem, (*role.Get())[index]); err != nil {
		return nil, err
	}

	return la, nil

}

type openaiLauncher struct {
	opt    *option
	client *openai.Client
	logger *zap.Logger
	io     ioer
}

func (ol *openaiLauncher) BackupHistory() error {
	err := ol.io.BackupData(ol.opt.userId, ol.opt.roleId)

	if err != nil {
		ol.logger.Error("Backup History error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return err
	}
	return nil
}

func (ol *openaiLauncher) SaveHistory(role, content string) error {
	newmsg := openai.ChatCompletionMessage{
		Role:    role,
		Content: content,
	}

	if flag := ol.io.IsExists(ol.opt.userId, ol.opt.roleId); !flag {
		temp := []openai.ChatCompletionMessage{}
		temp = append(temp, newmsg)

		bytes, err := json.Marshal(temp)
		if err != nil {
			ol.logger.Error("json Marshal error:", zap.String("error", fmt.Sprintf("%+v", err)))
			return err
		}
		if err := ol.io.SaveData(ol.opt.userId, ol.opt.roleId, bytes); err != nil {
			ol.logger.Error("WriteData error:", zap.String("error", fmt.Sprintf("%+v", err)))
			return err
		}
		return nil
	}

	//读取文件
	total := []openai.ChatCompletionMessage{}
	bytes, err := ol.io.ReadData(ol.opt.userId, ol.opt.roleId)
	if err != nil {
		ol.logger.Error("ReadData error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return err
	}

	if err := json.Unmarshal(bytes, &total); err != nil {
		ol.logger.Error("json Unmarshal error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return err
	}

	total = append(total, newmsg)
	bytes, err = json.Marshal(total)
	if err != nil {
		ol.logger.Error("json Marshal error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return err
	}
	if err := ol.io.SaveData(ol.opt.userId, ol.opt.roleId, bytes); err != nil {
		ol.logger.Error("WriteData error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return err
	}
	return nil

}

func (ol *openaiLauncher) assemble(role, content string) ([]openai.ChatCompletionMessage, error) {
	newmsg := openai.ChatCompletionMessage{
		Role:    role,
		Content: content,
	}

	//读取文件
	bytes, err := ol.io.ReadData(ol.opt.userId, ol.opt.roleId)
	if err != nil {
		ol.logger.Error("ReadData error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return nil, err
	}
	// 增加新聊天记录
	total := []openai.ChatCompletionMessage{}
	if err := json.Unmarshal(bytes, &total); err != nil {
		ol.logger.Error("json Unmarshal error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return nil, err
	}
	total = append(total, newmsg)
	return total, nil
}

func (ol *openaiLauncher) Chat(content string) (string, int, error) {
	// 组装对话
	array, err := ol.assemble(openai.ChatMessageRoleUser, content)
	if err != nil {
		ol.logger.Error("Chat Assemble error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return "", -1, err
	}

	resp, err := ol.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    ol.opt.chatModel,
			Messages: array,
		},
	)

	if err != nil {
		ol.logger.Error("ChatCompletion error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return "", -1, err
	}

	return resp.Choices[0].Message.Content, resp.Usage.TotalTokens, nil
}

func (ol *openaiLauncher) Translate(filepath string) (string, error) {
	resp, err := ol.client.CreateTranslation(
		context.Background(),
		openai.AudioRequest{
			Model:    ol.opt.audioModel,
			FilePath: filepath,
		},
	)
	if err != nil {
		ol.logger.Error("TranslateCompletion error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return "", err
	}

	return resp.Text, nil
}

func (ol *openaiLauncher) Moderations(content string) (bool, error) {
	resp, err := ol.client.Moderations(
		context.Background(),
		openai.ModerationRequest{
			Input: content,
		},
	)

	if err != nil {
		ol.logger.Error("Moderations error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return false, err
	}

	value := reflect.ValueOf(resp.Results[0].Categories)
	for i := 0; i < value.NumField(); i++ {
		if flag := value.Field(i).Bool(); flag {
			return false, nil
		}
	}

	return true, nil
}

func (ol *openaiLauncher) Close() error {
	return ol.BackupHistory()
}
