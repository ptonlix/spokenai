package manaudio

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ptonlix/spokenai/pkg/file"
	"github.com/ptonlix/spokenai/pkg/praudio"
	"github.com/ptonlix/spokenai/pkg/tts"
	"go.uber.org/zap"
)

const (
	recordFlag = 1
	playFlag   = 2
)

type Option func(*option)

type option struct {
	enablePlay bool
	audiohost  string
	//Chat Config
	roleId string
	userId string

	//Data Config
	dataRecordDir string
	dataPlayDir   string
}

func WithAudioHost(audiohost string) Option {
	return func(opt *option) {
		opt.audiohost = audiohost
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

func WithEnablePlay(enable bool) Option {
	return func(opt *option) {
		opt.enablePlay = enable
	}
}

type Manager struct {
	logger      *zap.Logger
	opt         *option
	recordCount int //该用户录制次数
	playCount   int //该用户的音频生成次数

	ttsclient *tts.TTSClient
}

func NewManager(logger *zap.Logger, options ...Option) (*Manager, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}
	// 初始化参数
	opt := new(option)
	for _, f := range options {
		f(opt)
	}
	c := tts.NewTTSClient(opt.audiohost)
	return &Manager{logger: logger, opt: opt, recordCount: 0, ttsclient: c}, nil

}
func (m *Manager) getListFilePath(flag int) []string {
	list := []string{}
	switch flag {
	case 1:
		for i := 1; i <= m.recordCount; i++ {
			list = append(list, m.opt.dataRecordDir+m.opt.userId+"_"+m.opt.roleId+"_"+strconv.Itoa(i)+".wav")
		}
	case 2:
		for i := 1; i <= m.playCount; i++ {
			list = append(list, m.opt.dataPlayDir+m.opt.userId+"_"+m.opt.roleId+"_"+strconv.Itoa(i)+".wav")
		}
	}

	return list
}

// 记录音频输入
func (m *Manager) RecordAudio(sig <-chan struct{}) {
	m.recordCount += 1
	filepath := m.GetRecordAudio()
	praudio.RecordAndSaveWithContext(context.Background(), filepath)
}

// 记录音频输入
func (m *Manager) RecordAudioWithContext(ctx context.Context, sig <-chan struct{}) {
	m.recordCount += 1
	filepath := m.GetRecordAudio()
	praudio.RecordAndSaveWithContext(ctx, filepath)
}

func (m *Manager) GetRecordAudio() string {
	return m.opt.dataRecordDir + m.opt.userId + "_" + m.opt.roleId + "_" + strconv.Itoa(m.recordCount) + ".wav"
}
func (m *Manager) GetPlayAudio() string {
	return m.opt.dataPlayDir + m.opt.userId + "_" + m.opt.roleId + "_" + strconv.Itoa(m.playCount) + ".wav"
}

// 播放音频文件
func (m *Manager) PlayAudio() error {
	if !m.opt.enablePlay {
		return nil
	}
	err := praudio.PlayWavFile(m.GetPlayAudio())
	if err != nil {
		m.logger.Error("Play AudioFile error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return err
	}
	return nil
}

// 生成音频文件
func (m *Manager) CallTTSserver(text string) error {
	if !m.opt.enablePlay {
		return nil
	}
	m.playCount += 1
	err := m.ttsclient.TextToSpeechSaveWav(text, m.GetPlayAudio())
	if err != nil {
		m.logger.Error("Call TTSserver error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return err
	}
	return nil
}

func (m *Manager) BackupRecordAudio() error {
	for _, filepath := range m.getListFilePath(recordFlag) {
		if _, ok := file.IsExists(filepath); !ok {
			m.logger.Warn("Backup RecordAudio File is not exist:", zap.String("filepath", filepath))
			continue
		} else {
			if err := praudio.BackupWAVfile(filepath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) BackupPlayAudio() error {
	for _, filepath := range m.getListFilePath(playFlag) {
		if _, ok := file.IsExists(filepath); !ok {
			m.logger.Warn("Backup PlayAudio File is not exist:", zap.String("filepath", filepath))
			continue
		} else {
			if err := praudio.BackupWAVfile(filepath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) Close() error {
	if err := m.BackupRecordAudio(); err != nil {
		m.logger.Warn("Backup RecordAudio error:", zap.String("error", fmt.Sprintf("%+v", err)))
		return err
	}
	if m.opt.enablePlay {
		if err := m.BackupPlayAudio(); err != nil {
			m.logger.Warn("Backup PlayAudio error:", zap.String("error", fmt.Sprintf("%+v", err)))
			return err
		}
	}
	return nil
}
