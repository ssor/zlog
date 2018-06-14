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

func (f *TextFormatter) Format(entry FormatterInput) ([]byte, error) {
    var b *bytes.Buffer
    var keys []string = make([]string, 0, len(entry.GetData()))
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
    fileInfo := formatShortFile()

    if isColored {
        f.printColored(b, entry, keys, timestampFormat, fileInfo)
    } else {
        if !f.DisableTimestamp {
            f.appendKeyValue(b, "time", entry.GetTime().Format(timestampFormat))
        }
        f.appendKeyValue(b, "", fileInfo)
        f.appendKeyValue(b, "level", entry.GetLevel().String())
        if entry.GetMessage() != "" {
            f.appendKeyValue(b, "msg", entry.GetMessage())
        }
        for _, key := range keys {
            //f.appendKeyValue(b, key, )
            value := fmt.Sprintf("%+v", entry.GetData()[key])
            if len(value) > 128 {
                value = value[:128] + "..."
            }
            fmt.Fprintf(b, "\n              - %-8s = %+v", key, value)
        }

        jsonRaw := entry.GetJsonRaw()
        if jsonRaw != nil {
            fmt.Fprintf(b, "\n%s", prettyJSON(jsonRaw))
        }
    }

    b.WriteByte('\n')
    return b.Bytes(), nil
}

func formatShortFile() string {
    _, file, line, ok := runtime.Caller(4)
    if !ok {
        file = "???"
        line = 0
        return "???:0"
    }

    short := ""
    for i := len(file) - 1; i > 0; i-- {
        if file[i] == '/' {
            short = file[i+1:]
            break
        }
    }
    return fmt.Sprintf("%s:%-3d", short, line)
}

func (f *TextFormatter) printColored(b *bytes.Buffer, entry FormatterInput, keys []string, timestampFormat, fileInfo string) {
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

    levelText := strings.ToUpper(entry.GetLevel().String())[0:4]

    if !f.FullTimestamp {
        fmt.Fprintf(b, "\x1b[%dm %s %s[%04d] %-44s \x1b[0m", levelColor, fileInfo, levelText, miniTS(), entry.GetMessage())
    } else {
        fmt.Fprintf(b, "\x1b[%dm %s %s[%s] %-44s \x1b[0m", levelColor, fileInfo, levelText, entry.GetTime().Format(timestampFormat), entry.GetMessage())
    }
    for _, k := range keys {
        value := fmt.Sprintf("%+v", entry.GetData()[k])
        if len(value) > 128 {
            value = value[:128] + "..."
        }
        fmt.Fprintf(b, "\n              \x1b[%dm- %-8s\x1b[0m = %+v", gray, k, value)
    }

    jsonRaw := entry.GetJsonRaw()
    if jsonRaw != nil {
        fmt.Fprintf(b, "\n%s", prettyJSON(jsonRaw))
    }
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
    var buf bytes.Buffer
    err := json.Indent(&buf, js, "", "  ")
    if err != nil {
        return "JSON parse error: " + err.Error()
    }
    s := string(buf.Bytes())
    return s
}
