package service

import (
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/config"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/public/cache"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/service/userMode"
	"strconv"
	"strings"
	"time"
	//"github.com/patrickmn/go-cache"
)

// UserServiceInterface 用户业务接口
type UserServiceInterface interface {
	GetUserSessionContext(userId string) string
	SetUserSessionContext(userId string, question, reply string)
	ClearUserSessionContext(userId string, msg string) bool
}

var _ UserServiceInterface = (*UserService)(nil)

// UserService 用戶业务
type UserService struct {
	// 缓存
	cache *cache.GptCache
}

// ClearUserSessionContext 清空GTP上下文，接收文本中包含 SessionClearToken
func (s *UserService) ClearUserSessionContext(userId string, msg string) bool {
	// 清空会话
	for _, clearToken := range config.LoadConfig().SessionClearTokens {
		if strings.Contains(msg, clearToken) {
			s.cache.Delete(userId)
			return true
		}
	}
	return false
}

// NewUserService 创建新的业务层
func NewUserService() UserServiceInterface {
	return &UserService{cache: cache.New(time.Second*config.LoadConfig().SessionTimeout, time.Minute*10)}
}

// GetUserSessionContext 获取用户会话上下文文本
func (s *UserService) GetUserSessionContext(userId string) string {

	sessionContext, ok := s.cache.Get(userId)
	if !ok {
		return ""
	}
	return sessionContext.(string)
}

// SetUserSessionContext 设置用户会话上下文文本，question用户提问内容，GTP回复内容
func (s *UserService) SetUserSessionContext(userId string, question, reply string) {
	value := question + "\n" + reply
	s.cache.Set(userId, value, time.Second*config.LoadConfig().SessionTimeout)
}

const userModeSuffix = "_USERMODE"

// GetUserMode  获取用户会话模式
func (s *UserService) GetUserMode(userId string) int {

	sessionContext, ok := s.cache.Get(userId + userModeSuffix)
	if !ok {
		return userMode.NotSet
	}
	ans, _ := strconv.Atoi(sessionContext.(string))
	return ans
}

// SetUserMode  设置用户会话模式
func (s *UserService) SetUserMode(userId string, mode int) {
	value := string(rune(mode))
	s.cache.Set(userId, value, time.Second*config.LoadConfig().SessionTimeout)
}
