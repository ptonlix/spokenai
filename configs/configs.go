package configs

import (
	"bytes"
	_ "embed"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ptonlix/spokenai/pkg/env"
	"github.com/ptonlix/spokenai/pkg/file"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var config = new(Config)

type Config struct {
	MySQL struct {
		Read struct {
			Addr string `toml:"addr"`
			User string `toml:"user"`
			Pass string `toml:"pass"`
			Name string `toml:"name"`
		} `toml:"read"`
		Write struct {
			Addr string `toml:"addr"`
			User string `toml:"user"`
			Pass string `toml:"pass"`
			Name string `toml:"name"`
		} `toml:"write"`
		Base struct {
			MaxOpenConn     int           `toml:"maxOpenConn"`
			MaxIdleConn     int           `toml:"maxIdleConn"`
			ConnMaxLifeTime time.Duration `toml:"connMaxLifeTime"`
		} `toml:"base"`
	} `toml:"mysql"`

	Redis struct {
		Addr         string `toml:"addr"`
		Pass         string `toml:"pass"`
		Db           int    `toml:"db"`
		MaxRetries   int    `toml:"maxRetries"`
		PoolSize     int    `toml:"poolSize"`
		MinIdleConns int    `toml:"minIdleConns"`
	} `toml:"redis"`

	Mail struct {
		Host string `toml:"host"`
		Port int    `toml:"port"`
		User string `toml:"user"`
		Pass string `toml:"pass"`
		To   string `toml:"to"`
	} `toml:"mail"`

	HashIds struct {
		Secret string `toml:"secret"`
		Length int    `toml:"length"`
	} `toml:"hashids"`

	Language struct {
		Local string `toml:"local"`
	} `toml:"language"`

	OpenAi struct {
		Chat struct {
			ChatModel       string `toml:"chatModel"`
			ChatMaxToken    int    `toml:"chatMaxtoken"`
			ChatTemperature int    `toml:"chatTemperature"`
			ChatTopP        int    `toml:"chatTopP"`
		} `toml:"chat"`
		Audio struct {
			AudioModel string `toml:"audioModel"`
		} `toml:"audio"`
		Base struct {
			ApiKey  string `toml:"apikey"`
			ApiHost string `toml:"apihost"`
		} `toml:"base"`
	} `toml:"openai"`

	File struct {
		History struct {
			Path string `toml:"path"`
		} `toml:"history"`
		Audio struct {
			Record struct {
				Path string `toml:"path"`
			} `toml:"record"`
			Play struct {
				Enable  bool   `toml:"enable"`
				Path    string `toml:"path"`
				TtsHost string `toml:"ttshost"`
			} `toml:"play"`
		} `toml:"audio"`
	}
}

var (
	//go:embed dev_configs.toml
	devConfigs []byte

	//go:embed fat_configs.toml
	fatConfigs []byte

	//go:embed uat_configs.toml
	uatConfigs []byte

	//go:embed pro_configs.toml
	proConfigs []byte
)

func init() {
	var r io.Reader

	switch env.Active().Value() {
	case "dev":
		r = bytes.NewReader(devConfigs)
	case "fat":
		r = bytes.NewReader(fatConfigs)
	case "uat":
		r = bytes.NewReader(uatConfigs)
	case "pro":
		r = bytes.NewReader(proConfigs)
	default:
		r = bytes.NewReader(fatConfigs)
	}

	viper.SetConfigType("toml")

	if err := viper.ReadConfig(r); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(config); err != nil {
		panic(err)
	}
	viper.SetConfigName(env.Active().Value() + "_configs")
	viper.AddConfigPath("./configs")

	configFile := "./configs/" + env.Active().Value() + "_configs.toml"
	_, ok := file.IsExists(configFile)
	if !ok {
		if err := os.MkdirAll(filepath.Dir(configFile), 0766); err != nil {
			panic(err)
		}

		f, err := os.Create(configFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if err := viper.WriteConfig(); err != nil {
			panic(err)
		}
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		if err := viper.Unmarshal(config); err != nil {
			panic(err)
		}
	})
}

func Get() Config {
	return *config
}
