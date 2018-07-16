package main

import "github.com/ssor/zlog"

func main() {

    obj := map[string]interface{}{
        "level1": map[string]interface{}{
            "level2":   "abc",
            "level2-1": 10000,
        },
        "level1-1": 10000,
    }

    zlog.Trace(0, zlog.Fields{"a1": "11111111111111111",}, obj, "trace 0")
    zlog.Trace(1, zlog.Fields{"a1": "11111111111111111",}, obj, "trace 0")
    zlog.Trace(1, zlog.Fields{"a1": "11111111111111111",}, obj, "trace 0")
    zlog.Trace(0, zlog.Fields{"a1": "11111111111111111",}, obj, "trace 0")

    zlog.TraceEnd(0)
    zlog.TraceEnd(1)
}
