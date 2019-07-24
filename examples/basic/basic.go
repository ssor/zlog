package main

import (
    "fmt"
    "github.com/ssor/zlog"
    "time"
)

func main() {
    go func() {
        ticker := time.NewTicker(3 * time.Second)
        for {
            <-ticker.C
            zlog.Info("ticking")
        }
    }()

    printTitle("NORMAL")
    fmt.Println("------ DebugLevel")
    zlog.SetLevel(zlog.DebugLevel)
    zlogPrint()
    fmt.Println("------ InfoLevel")
    zlog.SetLevel(zlog.InfoLevel)
    zlogPrint()
    fmt.Println("------ WarnLevel")
    zlog.SetLevel(zlog.WarnLevel)
    zlogPrint()
    fmt.Println("------ ErrorLevel")
    zlog.SetLevel(zlog.ErrorLevel)
    zlogPrint()

    printTitle("MODULE")
    log := zlog.New()
    fmt.Println("------ DebugLevel")
    log.SetLevel(zlog.DebugLevel)
    modulePrint(log)
    fmt.Println("------ InfoLevel")
    log.SetLevel(zlog.InfoLevel)
    modulePrint(log)
    fmt.Println("------ WarnLevel")
    log.SetLevel(zlog.WarnLevel)
    modulePrint(log)
    fmt.Println("------ ErrorLevel")
    log.SetLevel(zlog.ErrorLevel)
    modulePrint(log)

    printTitle("ChildLog")
    child := log.Sub("basic")
    child.Info("Info: This a child log")

    printTitle("DumpStacks")
    zlog.DumpStacks()

    printTitle("HIGHLIGHT")
    log.Highlight("this is highlight line with instance")
    log.Highlightf("this is highlightf line with instance")
    zlog.Highlight("Highlight: this is highlighted line too")
    zlog.Highlightf("Highlightf: this is highlighted line too")
}

func printTitle(title string) {
    fmt.Println()
    fmt.Println(fmt.Sprintf("------------------------------------------------ %s ------------------------------------------------", title))
    fmt.Println()
}

func zlogPrint() {
    zlog.Debug("debug test")
    zlog.Debugf("debug test")
    zlog.Info("info test")
    zlog.Infof("info test")
    zlog.Warn("warn test")
    zlog.Warnf("warn test")
    zlog.Error("Error test")
    zlog.Errorf("Error test")
    fmt.Println()
}

func modulePrint(log *zlog.Logger) {
    log.WithFields(zlog.Fields{
        "animal": "walrus",
        "number": 8,
    }).Debug("Debug: Started observing beach")

    log.WithFields(zlog.Fields{
        "animal": "walrus",
        "size":   10,
    }).Info("Info:   A group of walrus emerges from the ocean")

    log.WithFields(zlog.Fields{
        "omg":    true,
        "number": 122,
    }).Warn("Warn:   The group's number increased tremendously!")
    log.WithFields(zlog.Fields{
        "omg":    true,
        "number": 122,
    }).Error("Error: The group's number increased tremendously!")

    jsonRaw := []byte(`{"actions":{},"links":{},"message":"projects.management.cattle.io \"p-t4hzw\" is forbidden: User \"u-btx2s\" cannot get projects.management.cattle.io in the namespace \"c-g9v62\""}`)
    log.WithJsonRaw(jsonRaw).Info("info: try raw json print")

    obj := map[string]interface{}{
        "level1": map[string]interface{}{
            "level2":   "abc",
            "level2-1": 10000,
        },
        "level1-1": 10000,
    }
    log.WithStruct(obj).Debug("Debug: try struct print")
    log.WithStruct(obj).Info("Info:  try struct print")

    log.WithMultiLines("long", "例如，下面的命令输出一个 “Hello World”，之后终止容器。\n 例如，下面的命令输出一个 “Hello World”，之后终止容器。\n").Info("Info:  try long")

    log.WithFields(zlog.Fields{
        "animal": "orca",
        "size":   9009,
    }).Error("Error: It's over 9000!")
}
