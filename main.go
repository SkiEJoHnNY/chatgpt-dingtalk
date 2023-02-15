package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/config"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/public/logger"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/public/redlock"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/SkiEJoHnNY/chatgpt-dingtalk/gpt"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/public"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/service"
)

var UserService service.UserServiceInterface

func init() {
	UserService = service.NewUserService()
}

func main() {
	// log init
	logger.X()
	//
	Start()
}

func Start() {
	// 定义一个处理器函数
	handler := func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Warning("read request body failed: %v\n", err.Error())
			return
		}
		// TODO: 校验请求
		if len(data) == 0 {
			log.Warning("回调参数为空,以至于无法正常解析,请检查原因")
			return
		} else {
			var msgObj = new(public.ReceiveMsg)
			err = json.Unmarshal(data, &msgObj)
			if err != nil {
				log.Warning("unmarshal request body failed: %v\n", err)
			}
			log.Info(fmt.Sprintf("dingtalk callback parameters: %#v", msgObj))
			err = ProcessRequest(*msgObj)
			if err != nil {
				log.Warning("process request failed: %v\n", err)
			}
		}
	}

	// 创建一个新的 HTTP 服务器
	server := &http.Server{
		Addr:    ":8090",
		Handler: http.HandlerFunc(handler),
	}

	// 启动服务器
	log.Info("Start Listen On ", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Error(err)
	}
}

var expiration time.Duration = config.LoadConfig().ContextGapTime + 2*time.Second

func ProcessRequest(rmsg public.ReceiveMsg) error {
	atText := "@" + rmsg.SenderNick + "\n" + " "
	if UserService.ClearUserSessionContext(rmsg.SenderID, rmsg.Text.Content) {
		_, err := rmsg.ReplyText(atText + "上下文已经清空了，你可以问下一个问题啦。")
		if err != nil {
			log.Warning("response user error: %v \n", err)
			return err
		}
	} else {
		lock, err := redlock.Get(rmsg.SenderID, expiration, nil)
		log.Debug(lock, err)
		ctx := context.Background()
		if lock == nil {
			_, err = rmsg.ReplyText("宝 上一个回答还没返回呢")
			log.Debug(rmsg.SenderID, "Not Get RedisLock ")
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
			return nil
		}
		// 回复@我的用户
		reply = strings.TrimSpace(reply)
		reply = strings.Trim(reply, "\n")

		UserService.SetUserSessionContext(rmsg.SenderID, requestText, reply)
		lock.Release(ctx)
		replyText := atText + reply
		_, err = rmsg.ReplyText(replyText)
		if err != nil {
			log.Info("send message error: %v \n", err)
			return err
		}
	}
	return nil
}

// getRequestText 获取请求接口的文本，要做一些清洗
func getRequestText(rmsg public.ReceiveMsg) string {
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
	requestText = UserService.GetUserSessionContext(rmsg.SenderID) + requestText
	if len(requestText) >= 4000 {
		requestText = requestText[:4000]
	}

	// 4.返回请求文本
	return requestText
}
