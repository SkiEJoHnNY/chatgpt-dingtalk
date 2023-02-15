package logger

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
	"time"
)

var once sync.Once

func init() {
	once.Do(func() {
		writer1 := os.Stdout
		logPath := "/logs/chatgpt-dingtalk-x.log"
		fmt.Print("log filePath ", logPath)
		writer2, err := rotatelogs.New(logPath+".%Y%m%d%H%M", rotatelogs.WithLinkName(logPath),
			rotatelogs.WithMaxAge(time.Duration(72)*time.Hour),
			rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
		)
		if err != nil {
			log.Fatalf("create file log.txt failed: %v", err)
		}
		log.SetOutput(io.MultiWriter(writer1, writer2))
		log.SetLevel(log.DebugLevel)
	})
}

func X() {}
