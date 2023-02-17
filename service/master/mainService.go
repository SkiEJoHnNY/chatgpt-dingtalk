package master

import (
	"context"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/config"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/gpt"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/public/dingtalk"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/public/redlock"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/service/userMode"
	"sync"

	"github.com/SkiEJoHnNY/chatgpt-dingtalk/service/user"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type ServiceInterface interface {
	ProcessRequest(msg *dingtalk.ReceiveMsg) error
}
type Service struct {
}

var userService user.UserServiceInterface
var service ServiceInterface

func init() {
	userService = user.NewUserService()
}
func NewService() ServiceInterface {
	var once sync.Once
	once.Do(func() {
		service = &Service{}
	})
	return service
}

var expiration time.Duration = config.LoadConfig().ContextGapTime + 2*time.Second

func multiModeProcess(rmsg *dingtalk.ReceiveMsg) error {
	atText := "@" + rmsg.SenderNick + "\n" + " "
	lock, err := redlock.Get(rmsg.SenderID, expiration, nil)
	log.Debug(lock, err)
	ctx := context.Background()
	if lock == nil {
		_, err = rmsg.ReplyText("宝 上一个回答还没返回呢")
		log.Debug(rmsg.SenderID, " Not Get RedisLock ")
		return nil
	}
	lock.Refresh(ctx, expiration, nil)
	requestText := getRequestText(rmsg)
	// 获取问题的答案
	reply, err := gpt.Completions(requestText)
	if err != nil {
		log.Info("gpt request error: %v \n", err)
		_, err = rmsg.ReplyText("宝 机器人太累了，让她休息会儿，过一会儿再来请求。")
		if err != nil {
			log.Warning("send message error: %v \n", err)
			return err
		}
		log.Info("request openai error: %v\n", err)
		return err
	}
	if reply == "" {
		log.Warning("get gpt result falied: %v\n", err)
		_, err = rmsg.ReplyText(atText + "宝 openAI gpt 返回空string. 你的问题太蓝拉!")
		if err != nil {
			log.Warning("send message error: %v \n", err)
			return err
		}
		return nil
	}
	// 回复@我的用户
	reply = strings.TrimSpace(reply)
	reply = strings.Trim(reply, "\n")

	userService.SetUserSessionContext(rmsg.SenderID, requestText, reply)
	lock.Release(ctx)
	replyText := atText + reply
	_, err = rmsg.ReplyText(replyText)
	if err != nil {
		log.Info("send message error: %v \n", err)
		return err
	}
	return nil
}
func singleModeProcess(rmsg *dingtalk.ReceiveMsg) error {
	atText := "@" + rmsg.SenderNick + "\n" + " "
	requestText := rmsg.Text.Content
	// 获取问题的答案
	reply, err := gpt.Completions(requestText)
	if err != nil {
		log.Info("gpt request error: %v \n", err)
		_, err = rmsg.ReplyText("宝 机器人太累了，让她休息会儿，过一会儿再来请求。")
		if err != nil {
			log.Warning("send message error: %v \n", err)
			return err
		}
		log.Info("request openai error: %v\n", err)
		return err
	}
	if reply == "" {
		log.Warning("get gpt result falied: %v\n", err)
		_, err = rmsg.ReplyText(atText + "宝 openAI gpt 返回空string. 你的问题太蓝拉!")
		if err != nil {
			log.Warning("send message error: %v \n", err)
			return err
		}
		return nil
	}
	// 回复@我的用户
	reply = strings.TrimSpace(reply)
	reply = strings.Trim(reply, "\n")
	replyText := atText + reply
	_, err = rmsg.ReplyText(replyText)
	if err != nil {
		log.Info("send message error: %v \n", err)
		return err
	}
	return nil
}
func (s *Service) ProcessRequest(rmsg *dingtalk.ReceiveMsg) error {
	atText := "@" + rmsg.SenderNick + "\n" + " "
	userId := rmsg.SenderID

	if !userService.UserCommand(rmsg) {
		usermode := userService.GetUserMode(userId)
		if usermode == userMode.NotSet {
			userService.SetUserMode(userId, userMode.SingleQA)
			rmsg.ReplyText(atText + " 默认设置成单聊模式")
		}
		switch usermode {
		case userMode.SingleQA:
			return singleModeProcess(rmsg)
		case userMode.MultiQA:
			return multiModeProcess(rmsg)
		}
	}
	return nil
}

// getRequestText 获取请求接口的文本，要做一些清洗
func getRequestText(rmsg *dingtalk.ReceiveMsg) string {
	// 1.去除空格以及换行
	requestText := strings.TrimSpace(rmsg.Text.Content)
	requestText = strings.Trim(rmsg.Text.Content, "\n")
	// 2.替换掉当前用户名称
	replaceText := "@" + rmsg.SenderNick
	requestText = strings.TrimSpace(strings.ReplaceAll(rmsg.Text.Content, replaceText, ""))
	if requestText == "" {
		return ""
	}

	// 3.获取上下文，拼接在一起，如果字符长度超出4000，截取为4000。（GPT按字符长度算）
	requestText = userService.GetUserSessionContext(rmsg.SenderID) + requestText
	if len(requestText) >= 4000 {
		requestText = requestText[:4000]
	}

	// 4.返回请求文本
	return requestText
}
