package user

import (
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/config"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/public/cache"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/public/dingtalk"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/service/userMode"
	"strconv"
	"strings"
	"sync"
	"time"
	//"github.com/patrickmn/go-cache"
)

// UserServiceInterface 用户业务接口
type UserServiceInterface interface {
	GetUserSessionContext(userId string) string
	SetUserSessionContext(userId string, question, reply string)
	ClearUserSessionContext(userId string, msg string) bool
	UserCommand(*dingtalk.ReceiveMsg) bool
	GetUserMode(string) int
	SetUserMode(string, int)
}

var _ UserServiceInterface = (*UserService)(nil)

// UserService 用戶业务
type UserService struct {
	// 缓存
	cache *cache.GptCache
}

var welcomeStr string = " 你好你好,我是小d. 我是一个笨蛋,我不能给你提供帮助哦!"

// UserCommand 用户命令接口  非命令返回false
func (s *UserService) UserCommand(rmsg *dingtalk.ReceiveMsg) bool {
	userId := rmsg.SenderID
	msg := rmsg.Text.Content
	atText := "@" + rmsg.SenderNick + "\n" + " "
	switch msg {
	case "单聊":
		s.SetUserMode(userId, userMode.SingleQA)
		rmsg.ReplyText(atText + " 设置单聊模式成功")
	case "串聊":
		s.SetUserMode(userId, userMode.MultiQA)
		rmsg.ReplyText(atText + " 设置串聊模式成功")
	case "help":
		rmsg.ReplyText(welcomeStr)
	default:
		if s.ClearUserSessionContext(userId, msg) {
			rmsg.ReplyText(atText + "上下文已经清空了，你可以问下一个问题啦。")
			return true
		} else {
			return false
		}
	}
	return true
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

var once sync.Once
var userService UserServiceInterface

// NewUserService 创建新的业务层
func NewUserService() UserServiceInterface {
	once.Do(func() {
		userService = &UserService{cache: cache.New(time.Second*config.LoadConfig().SessionTimeout, time.Minute*10)}
	})
	return userService
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
