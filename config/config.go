package config

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Configuration 项目配置
// ContextGapTime SessionTimeout seconds
type Configuration struct {
	// gtp apikey
	ApiKey string `json:"api_key"`
	// 会话超时时间
	SessionTimeout time.Duration `json:"session_timeout"`
	// GPT请求最大字符数
	MaxTokens uint `json:"max_tokens"`
	// GPT模型
	Model string `json:"model"`
	// 热度
	Temperature float64 `json:"temperature"`
	// 自定义清空会话口令
	SessionClearTokenStrings string `json:"session_clear_token"`
	SessionClearTokens       []string
	ContextGapTime           time.Duration `json:"context_gap_time"`
}

var config *Configuration
var once sync.Once

// LoadConfig 加载配置
func LoadConfig() *Configuration {
	once.Do(func() {
		// 从文件中读取
		config = &Configuration{}
		f, err := os.Open("config.json")
		if err != nil {
			log.Error("open config err: %v", err)
			return
		}
		defer f.Close()
		encoder := json.NewDecoder(f)
		err = encoder.Decode(config)
		if err != nil {
			log.Warning("decode config err: %v", err)
			return
		}

		// 如果环境变量有配置，读取环境变量
		ApiKey := os.Getenv("APIKEY")
		SessionTimeout := os.Getenv("SESSION_TIMEOUT")
		Model := os.Getenv("MODEL")
		MaxTokens := os.Getenv("MAX_TOKENS")
		Temperature := os.Getenv("TEMPERATURE")
		SessionClearToken := os.Getenv("SESSION_CLEAR_TOKEN")
		if ApiKey != "" {
			config.ApiKey = ApiKey
		}
		if SessionTimeout != "" {
			duration, err := strconv.ParseInt(SessionTimeout, 10, 64)
			if err != nil {
				log.Error(fmt.Sprintf("config session timeout err: %v ,get is %v", err, SessionTimeout))
				return
			}
			config.SessionTimeout = time.Duration(duration) * time.Second
		} else {
			config.SessionTimeout = time.Duration(config.SessionTimeout) * time.Second
		}
		if Model != "" {
			config.Model = Model
		}
		if MaxTokens != "" {
			max, err := strconv.Atoi(MaxTokens)
			if err != nil {
				log.Error(fmt.Sprintf("config MaxTokens err: %v ,get is %v", err, MaxTokens))
				return
			}
			config.MaxTokens = uint(max)
		}
		if Temperature != "" {
			temp, err := strconv.ParseFloat(Temperature, 64)
			if err != nil {
				log.Error(fmt.Sprintf("config Temperature err: %v ,get is %v", err, Temperature))
				return
			}
			config.Temperature = temp
		}
		if SessionClearToken != "" {
			config.SessionClearTokenStrings = SessionClearToken
		}
		config.ContextGapTime = config.ContextGapTime * time.Second
		config.SessionClearTokens = strings.Split(config.SessionClearTokenStrings, ",")
		log.Info("config.SessionClearTokens is ", config.SessionClearTokens)
		log.Debug(config)
	})
	if config.ApiKey == "" {
		log.Error("config err: api key required")
	}
	return config
}
