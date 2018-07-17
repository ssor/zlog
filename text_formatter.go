package zlog

import (
    "bytes"
    "fmt"
    "runtime"
    "sort"
    "strings"
    "time"
    "encoding/json"
)

const (
    nocolor = 0
    red     = 31
    green   = 32
    yellow  = 33
    blue    = 34
    gray    = 37
)

var (
    baseTimestamp time.Time
    isTerminal    bool
)

func init() {
    baseTimestamp = time.Now()
    isTerminal = IsTerminal()
}

func miniTS() int {
    return int(time.Since(baseTimestamp) / time.Second)
}

type TextFormatter struct {
    // Set to true to bypass checking for a TTY before outputting colors.
    ForceColors bool

    // Force disabling colors.
    DisableColors bool

    // Disable timestamp logging. useful when output is redirected to logging
    // system that already adds timestamps.
    DisableTimestamp bool

    // Enable logging the full timestamp when a TTY is attached instead of just
    // the time passed since beginning of execution.
    FullTimestamp bool

    // TimestampFormat to use for display when a full timestamp is printed
    TimestampFormat string

    // The fields are sorted by default for a consistent output. For applications
    // that log extremely frequently and don't use the JSON formatter this may not
    // be desired.
    DisableSorting bool
}

func dumpStacks() {

    prefix := "      "
    buf := make([]byte, 16384)
    buf = buf[:runtime.Stack(buf, true)]
    results := bytes.Split(buf, []byte{0xa})
    //spew.Dump(results)
    if len(results) > 6 {
        fmt.Println(prefix + "=== BEGIN goroutine stack dump ===")
        for i := 6; i < len(results); i += 1 {
            if len(results[i]) <= 0 { //do not print all goroutine stacks
                break
            }
            if i%2 == 0 { //just print code line
                content := strings.Replace(string(results[i]), gopath, "", 1)
                fmt.Printf("%s%s: %s\n", prefix, "- ", strings.TrimSpace(content))
            }
        }
        fmt.Println(prefix + "=== END goroutine stack dump ===")
    } else {
        fmt.Println(prefix + "=== no stack to print ===")
    }
}

func (f *TextFormatter) Format(entry FormatterInput, callDepth int) ([]byte, error) {
    var b *bytes.Buffer
    var keys = make([]string, 0, len(entry.GetData()))
    for k := range entry.GetData() {
        keys = append(keys, k)
    }

    if !f.DisableSorting {
        sort.Strings(keys)
    }
    if entry.GetBuffer() != nil {
        b = entry.GetBuffer()
    } else {
        b = &bytes.Buffer{}
    }

    prefixFieldClashes(entry.GetData())

    isColorTerminal := isTerminal && (runtime.GOOS != "windows")
    isColored := (f.ForceColors || isColorTerminal) && !f.DisableColors

    timestampFormat := f.TimestampFormat
    if timestampFormat == "" {
        timestampFormat = DefaultTimestampFormat
    }
    //fileInfo := formatShortFile(callDepth)
    //isColored = false
    if isColored {
        f.printColored(b, entry, keys, timestampFormat)
    } else {
        fmt.Fprintf(b, "%s%-44s  (%s)[%s]", entry.GetLevel().String(), entry.GetMessage(), entry.GetData()[moduleKey], entry.GetTime().Format(timestampFormat))

        for _, key := range keys {
            if key == moduleKey {
                continue
            }
            //f.appendKeyValue(b, key, )
            value := fmt.Sprintf("%+v", entry.GetData()[key])
            fmt.Fprintf(b, "\n     - %-8s = %+v", key, tripHeadAndTail(value, 128))
        }

        jsonRaw := entry.GetJsonRaw()
        if jsonRaw != nil {
            fmt.Fprintf(b, "\n%s", prettyJSON(jsonRaw))
        }
    }

    b.WriteByte('\n')
    return b.Bytes(), nil
}

//
//func formatShortFile(callDepth int) string {
//    _, file, line, ok := runtime.Caller(callDepth)
//    if !ok {
//        file = "???"
//        line = 0
//        return "???:0"
//    }
//
//    short := ""
//    for i := len(file) - 1; i > 0; i-- {
//        if file[i] == '/' {
//            short = file[i+1:]
//            break
//        }
//    }
//    //DumpStacks()
//    return fmt.Sprintf(" [%s:%-3d]", short, line)
//}

func (f *TextFormatter) printColored(b *bytes.Buffer, entry FormatterInput, keys []string, timestampFormat string) {
    var levelColor int
    switch entry.GetLevel() {
    case DebugLevel:
        levelColor = gray
    case WarnLevel:
        levelColor = yellow
    case ErrorLevel, FatalLevel, PanicLevel:
        levelColor = red
    default:
        levelColor = blue
    }

    levelText := strings.ToUpper(entry.GetLevel().String())

    if !f.FullTimestamp {
        fmt.Fprintf(b, "\x1b[%dm %s%-44s  (%s)[%04d]\x1b[0m", levelColor, levelText, entry.GetMessage(), entry.GetData()[moduleKey], miniTS())
    } else {
        fmt.Fprintf(b, "\x1b[%dm %s %-44s  (%s)[%s]\x1b[0m", levelColor, levelText, entry.GetMessage(), entry.GetData()[moduleKey], entry.GetTime().Format(timestampFormat))
    }
    for _, k := range keys {
        value := fmt.Sprintf("%+v", entry.GetData()[k])
        fmt.Fprintf(b, "\n      \x1b[%dm- %-8s = %+v \x1b[0m", gray, k, tripHeadAndTail(value, 128))
    }

    jsonRaw := entry.GetJsonRaw()
    if jsonRaw != nil {
        fmt.Fprintf(b, "\x1b[%dm \n%s \x1b[0m", gray, prettyJSON(jsonRaw))
    }
}

func tripHeadAndTail(src string, count int) string {
    length := len(src)
    if length <= count {
        return src
    }

    if count%2 != 0 {
        count ++
    }
    return src[:count/2] + "..." + src[length-count/2:length-1]
}

func needsQuoting(text string) bool {
    for _, ch := range text {
        if !((ch >= 'a' && ch <= 'z') ||
            (ch >= 'A' && ch <= 'Z') ||
            (ch >= '0' && ch <= '9') ||
            ch == '-' || ch == '.') {
            return true
        }
    }
    return false
}

func (f *TextFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
    if len(key) > 0 {
        b.WriteString(key)
        b.WriteByte('=')
    }

    switch value := value.(type) {
    case string:
        if !needsQuoting(value) {
            b.WriteString(value)
        } else {
            fmt.Fprintf(b, "%q", value)
        }
    case error:
        errmsg := value.Error()
        if !needsQuoting(errmsg) {
            b.WriteString(errmsg)
        } else {
            fmt.Fprintf(b, "%q", value)
        }
    default:
        fmt.Fprint(b, value)
    }

    b.WriteByte(' ')
}

func prettyJSON(js []byte) string {
    prefix := "      "
    var buf bytes.Buffer
    err := json.Indent(&buf, js, prefix, "  ")
    if err != nil {
        return "JSON parse error: " + err.Error()
    }
    s := string(buf.Bytes())
    return prefix + s
}
