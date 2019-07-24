package main

import (
    "fmt"
    "github.com/ssor/zlog"
    "gopkg.in/natefinch/lumberjack.v2"
    "io"
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

    for _, name := range []string{"module1", "module2", "module3"} {
        go startLogger(name, logFile)
    }
    select {}
}

func startLogger(name string, writer io.Writer) {
    logger := zlog.New(name)
    logger.SetOutput(writer)

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
            fmt.Println(name, " -> now is ", now)
        }
        logger.Infof("now is %s", now.String())
    }
}
