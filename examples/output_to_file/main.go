package main

import (
	"fmt"
	"github.com/ssor/zlog"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

func main() {

	logFile := &lumberjack.Logger{
		Filename:   "main.log",
		MaxSize:    1, // megabytes
		MaxBackups: 3,
		MaxAge:     28,    //days
		Compress:   false, // disabled by default
	}

	var loggers []*zlog.Logger
	for _, name := range []string{"module1", "module2", "module3"} {
		loggers = append(loggers, zlog.New(name))
	}

	zlog.SetOutput(logFile)

	for _, logger := range loggers {
		go startLogger(logger)
	}
	select {}
}

func startLogger(logger *zlog.Logger) {
	start := time.Now()
	duration := 5 * time.Minute
	count := 0
	for {
		now := time.Now()
		if now.Sub(start) > duration {
			break
		}
		count++
		time.Sleep(1 * time.Millisecond)
		if count%1000 == 0 {
			fmt.Println(logger.Name(), " -> now is ", now)
		}
		logger.Infof("now is %s", now.String())
	}
}
