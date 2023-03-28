package clean

import (
	"errors"

	"github.com/ptonlix/spokenai/pkg/file"
	"go.uber.org/zap"
)

type Cleaner interface {
	ClearAllData() error
}

type Option func(*option)

type option struct {
	//Data Config
	enableConsole bool
	dataChatDir   string

	//Data Config
	dataRecordDir string
	dataPlayDir   string
}

func WithEnableConsole() Option {
	return func(opt *option) {
		opt.enableConsole = true
	}
}

func WithDataChatDir(path string) Option {
	return func(opt *option) {
		opt.dataChatDir = path
	}
}

func WithDataRecordDir(path string) Option {
	return func(opt *option) {
		opt.dataRecordDir = path
	}
}

func WithDataPlayDir(path string) Option {
	return func(opt *option) {
		opt.dataPlayDir = path
	}
}

type ConsoleClean struct {
	logger      *zap.Logger
	dataChatDir string

	//Data Config
	dataRecordDir string
	dataPlayDir   string
	err           error
}

func NewCleaner(logger *zap.Logger, options ...Option) (Cleaner, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}
	// 初始化参数
	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	return &ConsoleClean{logger: logger, dataChatDir: opt.dataChatDir, dataRecordDir: opt.dataRecordDir, dataPlayDir: opt.dataPlayDir}, nil
}

func (c *ConsoleClean) GetErr() error {
	return c.err
}

func (c *ConsoleClean) ClearChatData() {
	if c.err != nil {
		return
	}
	if err := file.RemoveContents(c.dataChatDir); err != nil {
		c.err = err
	}
}

func (c *ConsoleClean) ClearRecordData() {
	if c.err != nil {
		return
	}
	if err := file.RemoveContents(c.dataRecordDir); err != nil {
		c.err = err
	}
}

func (c *ConsoleClean) ClearPlayData() {
	if c.err != nil {
		return
	}
	if err := file.RemoveContents(c.dataPlayDir); err != nil {
		c.err = err
	}
}

func (c *ConsoleClean) ClearAllData() error {
	c.ClearChatData()
	c.ClearPlayData()
	c.ClearRecordData()
	return c.err
}
