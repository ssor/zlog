package main

import (
    "github.com/ssor/zlog"
)

var log = zlog.New()

func init() {
    //log.Formatter = new(zlog.JSONFormatter)
    log.Formatter = new(zlog.TextFormatter) // default
    log.Level = zlog.DebugLevel
}

func main() {

    defer func() {
        err := recover()
        if err != nil {
            log.WithFields(zlog.Fields{
                "omg":    true,
                "err":    err,
                "number": 100,
            }).Fatal("The ice breaks!")
        }
    }()

    zlog.SetLevel(zlog.DebugLevel)
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

    log.WithFields(zlog.Fields{
        "animal": "orca",
        "size":   9009,
    }).Panic("It's over 9000!")
}
