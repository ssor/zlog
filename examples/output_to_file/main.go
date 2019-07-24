package main

import (
    "fmt"
    "github.com/ssor/zlog"
    "gopkg.in/natefinch/lumberjack.v2"
    "time"
)

func main() {

    name := "main"
    logger := zlog.New()
    logger.SetOutput(&lumberjack.Logger{
        Filename:   name + ".log",
        MaxSize:    1, // megabytes
        MaxBackups: 3,
        MaxAge:     28,    //days
        Compress:   false, // disabled by default
    })

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
            fmt.Println("now is ", now)
        }
        logger.Infof("now is %s", now.String())
    }
}
