package main

import (
	"encoding/json"
	"fmt"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/public/dingtalk"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/public/logger"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/service/master"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

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
			var msgObj = new(dingtalk.ReceiveMsg)
			err = json.Unmarshal(data, &msgObj)
			if err != nil {
				if err != nil {
					log.Warning("unmarshal request body failed: %v\n", err)
				}
				log.Info(fmt.Sprintf("dingtalk callback parameters: %#v", msgObj))
				err = master.NewService().ProcessRequest(msgObj)
				if err != nil {
					log.Warning("process request failed: %v\n", err)
				}
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
