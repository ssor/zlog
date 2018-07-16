package main

import (
    "github.com/ssor/zlog"
    "time"
)

var log = zlog.New()

func init() {
    //log.Formatter = new(zlog.JSONFormatter)
    log.Formatter = new(zlog.TextFormatter) // default
    log.Level = zlog.DebugLevel
}

func main() {
    go func() {
        ticker := time.NewTicker(3 * time.Second)
        for {
            <-ticker.C
            zlog.Info("ticking")
        }
    }()

    zlog.SetLevel(zlog.DebugLevel)
    zlog.SetLevel(zlog.InfoLevel)
    zlog.Debug("debug test")
    zlog.Info("info test")

    log.WithFields(zlog.Fields{
        "animal": "walrus",
        "number": 8,
    }).Debug("Started observing beach")

    log.WithFields(zlog.Fields{
        "animal": "walrus",
        "size":   10,
    }).Info("A group of walrus emerges from the ocean")

    log.WithFields(zlog.Fields{
        "omg":    true,
        "number": 122,
    }).Warn("The group's number increased tremendously!")

    log.WithFields(zlog.Fields{
        "temperature": -4,
    }).Debug("Temperature changes")

    jsonRaw := []byte(`{"actions":{},"baseType":"error","code":"Forbidden","links":{},"message":"projects.management.cattle.io \"p-t4hzw\" is forbidden: User \"u-btx2s\" cannot get projects.management.cattle.io in the namespace \"c-g9v62\"","status":403,"type":"error"}`)
    log.WithJsonRaw(jsonRaw).Debug("try json print")

    obj := map[string]interface{}{
        "level1": map[string]interface{}{
            "level2":   "abc",
            "level2-1": 10000,
        },
        "level1-1": 10000,
    }
    log.WithStruct(obj).Debug("try struct print")
    log.WithStruct(obj).Info("try struct print")

    log.WithMultiLines("long", "例如，下面的命令输出一个 “Hello World”，之后终止容器。\n 例如，下面的命令输出一个 “Hello World”，之后终止容器。\n").Info("try long")

    zlog.DumpStacks()

    log.WithFields(zlog.Fields{
        "animal": "orca",
        "size":   9009,
    }).Panic("It's over 9000!")

    child := log.Sub("basic")
    child.Info("This a child log")
}
