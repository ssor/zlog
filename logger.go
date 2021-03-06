package zlog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
)

var loggerDefaultCallDepth = 6

type Logger struct {
	// The logs are `io.Copy`'d to this in a mutex. It's common to set this to a
	// file, or leave it default which is `os.Stderr`. You can also set this to
	// something more adventorous, such as logging to Kafka.
	Out io.Writer
	// All log entries pass through the formatter before logged to Out. The
	// included formatters are `TextFormatter` and `JSONFormatter` for which
	// TextFormatter is the default. In development (when a TTY is attached) it
	// logs with colors, but to a file it wouldn't. You can easily implement your
	// own that implements the `Formatter` interface, see the `README` or included
	// formatters for examples.
	Formatter Formatter
	// The logging level the logger should log at. This is typically (and defaults
	// to) `logrus.Info`, which allows Info(), Warn(), Error() and Fatal() to be
	// logged. `logrus.Debug` is useful in
	Level Level
	// Used to sync writing to the log. Locking is enabled by Default
	mu MutexWrap
	// Reusable empty entry
	entryPool sync.Pool

	moduleName string
}

type MutexWrap struct {
	lock     sync.Mutex
	disabled bool
}

func (mw *MutexWrap) Lock() {
	if !mw.disabled {
		mw.lock.Lock()
	}
}

func (mw *MutexWrap) Unlock() {
	if !mw.disabled {
		mw.lock.Unlock()
	}
}

func (mw *MutexWrap) Disable() {
	mw.disabled = true
}

// Creates a new logger. Configuration should be set by changing `Formatter`,
// `Out` and `Hooks` directly on the default logger instance. You can also just
// instantiate your own:
//
//    var log = &Logger{
//      Out: os.Stderr,
//      Formatter: new(JSONFormatter),
//      Hooks: make(LevelHooks),
//      Level: logrus.DebugLevel,
//    }
//
// It's recommended to make this a global instance called `log`.
func New(moduleNames ...string) *Logger {
	if moduleNames == nil || len(moduleNames) <= 0 {
		moduleNames = []string{"main"}
	}

	logger := &Logger{
		Out:        os.Stderr,
		Formatter:  new(TextFormatter),
		Level:      DebugLevel,
		moduleName: strings.Join(moduleNames, "/"),
	}
	loggers = append(loggers, logger)
	return logger
}

func (logger *Logger) Name() string {
	return logger.moduleName
}

// SetOutput sets the standard logger output.
func (logger *Logger) SetOutput(out io.Writer) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.Out = out
}

func (logger *Logger) SetLevel(level Level) {
	logger.Level = level
}

func (logger *Logger) newEntry() *Entry {
	entry, ok := logger.entryPool.Get().(*Entry)
	if ok {
		return entry
	}
	return NewEntry(logger, logger.moduleName)
}

func (logger *Logger) releaseEntry(entry *Entry) {
	logger.entryPool.Put(entry)
}

func (logger *Logger) WithStruct(value interface{}) *Entry {
	bs, err := json.Marshal(value)
	if err != nil {
		return logger.WithError(err)
	}
	return logger.WithJsonRaw(bs)
}

func (logger *Logger) WithJsonRaw(bs []byte) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithJsonRaw(bs)
}

func (logger *Logger) WithMultiLines(key, longStr string) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	fields := make(Fields)
	lns := strings.Split(longStr, "\n")
	for index, ln := range lns {
		if len(ln) <= 0 {
			continue
		}
		fields[fmt.Sprintf("%s-%d", key, index)] = ln
	}
	return entry.WithFields(fields)
}

func (logger *Logger) WithLongString(key, longStr, sep string) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	fields := make(Fields)
	lns := strings.Split(longStr, sep)
	for index, ln := range lns {
		if len(ln) <= 0 {
			continue
		}
		fields[fmt.Sprintf("%s-%d", key, index)] = ln
	}
	return entry.WithFields(fields)
}

// Adds a field to the log entry, note that it doesn't log until you call
// Debug, Print, Info, Warn, Fatal or Panic. It only creates a log entry.
// If you want multiple fields, use `WithFields`.
func (logger *Logger) WithField(key string, value interface{}) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithField(key, value)
}

// Adds a struct of fields to the log entry. All it does is call `WithField` for
// each `Field`.
func (logger *Logger) WithFields(fields Fields) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithFields(fields)
}

// Add an error as single field to the log entry.  All it does is call
// `WithError` for the given `error`.
func (logger *Logger) WithError(err error) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithError(err)
}

type LevelParser interface {
	Parse(*string) (Level, error)
}

type RegexpParser struct {
	r *regexp.Regexp
}

func (pr *RegexpParser) prefixRegex() {
	//pr.r = regexp.MustCompile(`^\\[(?P<Level?\\w+)\\]`)
	pr.r = regexp.MustCompile(`^\[\w+\]`)
}

func (pr *RegexpParser) Parse(s *string) (Level, error) {
	b := pr.r.Find([]byte(*s))
	return ParseLevel(string(b)[1 : len(b)-1])
}

type PrefixStrCmp struct{}

func (p *PrefixStrCmp) Parse(s *string) (Level, error) {
	str := *s
	prefix := str[:7]

	switch prefix {
	case "[INFO] ":
		return InfoLevel, nil
	case "[WARN] ":
		return WarnLevel, nil
	case "[ERROR]":
		return ErrorLevel, nil
	case "[FATAL]":
		return FatalLevel, nil
	case "[DEBUG]":
		return DebugLevel, nil
	case "[PANIC]":
		return PanicLevel, nil
	default:
		return DebugLevel, nil
	}
	return WarnLevel, fmt.Errorf("prefixstrcmp switch failed?")
}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	if logger.Level >= DebugLevel {
		entry := logger.newEntry()
		//entry.Debugf(format, args...)
		entry.log(0, DebugLevel, fmt.Sprintf(format, args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	if logger.Level >= InfoLevel {
		entry := logger.newEntry()
		entry.log(0, InfoLevel, fmt.Sprintf(format, args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Printf(format string, args ...interface{}) {
	entry := logger.newEntry()
	entry.log(0, logger.Level, fmt.Sprintf(format, args...))
	logger.releaseEntry(entry)
}

func (logger *Logger) Highlightf(format string, args ...interface{}) {
	logger.highlight(loggerDefaultCallDepth, fmt.Sprintf(format, args...))
}

func (logger *Logger) Highlight(args ...interface{}) {
	logger.highlight(loggerDefaultCallDepth, args...)
}

func (logger *Logger) highlight(callDepth int, args ...interface{}) {
	entry := logger.newEntry()
	entry.log(callDepth, ErrorLevel, fmt.Sprint(args...))
	logger.releaseEntry(entry)
}

func (logger *Logger) Warnf(format string, args ...interface{}) {
	if logger.Level >= WarnLevel {
		entry := logger.newEntry()
		entry.log(0, WarnLevel, fmt.Sprintf(format, args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Warningf(format string, args ...interface{}) {
	if logger.Level >= WarnLevel {
		entry := logger.newEntry()
		entry.log(0, WarnLevel, fmt.Sprintf(format, args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	if logger.Level >= ErrorLevel {
		entry := logger.newEntry()
		entry.log(0, ErrorLevel, fmt.Sprintf(format, args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Fatalf(format string, args ...interface{}) {
	if logger.Level >= FatalLevel {
		entry := logger.newEntry()
		entry.log(0, FatalLevel, fmt.Sprintf(format, args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Panicf(format string, args ...interface{}) {
	if logger.Level >= PanicLevel {
		entry := logger.newEntry()
		entry.log(0, PanicLevel, fmt.Sprintf(format, args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Debug(args ...interface{}) {
	if logger.Level >= DebugLevel {
		entry := logger.newEntry()
		entry.log(0, DebugLevel, fmt.Sprint(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Passf(format string, args ...interface{}) {
	entry := logger.newEntry()
	entry.log(0, InfoLevel, fmt.Sprintf("[PASS]"+format, args...))
	logger.releaseEntry(entry)
}

func (logger *Logger) Pass(args ...interface{}) {
	entry := logger.newEntry()
	args = append([]interface{}{"[PASS]"}, args...)
	entry.log(0, InfoLevel, fmt.Sprint(args...))
	logger.releaseEntry(entry)
}
func (logger *Logger) Failedf(format string, args ...interface{}) {
	entry := logger.newEntry()
	entry.log(0, ErrorLevel, fmt.Sprintf("[FAIL]"+format, args...))
	logger.releaseEntry(entry)
}

func (logger *Logger) Failed(args ...interface{}) {
	entry := logger.newEntry()
	args = append([]interface{}{"[FAIL]"}, args...)
	entry.log(0, ErrorLevel, fmt.Sprint(args...))
	logger.releaseEntry(entry)
}
func (logger *Logger) Successf(format string, args ...interface{}) {
	entry := logger.newEntry()
	entry.log(0, InfoLevel, fmt.Sprintf("[OK]"+format, args...))
	logger.releaseEntry(entry)
}

func (logger *Logger) Success(args ...interface{}) {
	entry := logger.newEntry()
	args = append([]interface{}{"[OK]"}, args...)
	entry.log(0, InfoLevel, fmt.Sprint(args...))
	logger.releaseEntry(entry)
}

func (logger *Logger) Info(args ...interface{}) {
	if logger.Level >= InfoLevel {
		entry := logger.newEntry()
		entry.log(0, InfoLevel, fmt.Sprint(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Print(args ...interface{}) {
	entry := logger.newEntry()
	entry.log(0, logger.Level, fmt.Sprint(args...))
	logger.releaseEntry(entry)
}

func (logger *Logger) Warn(args ...interface{}) {
	if logger.Level >= WarnLevel {
		entry := logger.newEntry()
		entry.log(0, WarnLevel, fmt.Sprint(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Error(args ...interface{}) {
	if logger.Level >= ErrorLevel {
		entry := logger.newEntry()
		entry.log(0, ErrorLevel, fmt.Sprint(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Fatal(args ...interface{}) {
	if logger.Level >= FatalLevel {
		entry := logger.newEntry()
		entry.log(0, FatalLevel, fmt.Sprint(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Panic(args ...interface{}) {
	if logger.Level >= PanicLevel {
		entry := logger.newEntry()
		entry.log(0, PanicLevel, fmt.Sprint(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Debugln(args ...interface{}) {
	if logger.Level >= DebugLevel {
		entry := logger.newEntry()
		entry.log(0, DebugLevel, fmt.Sprintln(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Infoln(args ...interface{}) {
	if logger.Level >= InfoLevel {
		entry := logger.newEntry()
		entry.log(0, InfoLevel, fmt.Sprintln(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Println(args ...interface{}) {
	entry := logger.newEntry()
	entry.log(0, logger.Level, fmt.Sprintln(args...))
	logger.releaseEntry(entry)
}

func (logger *Logger) Warnln(args ...interface{}) {
	if logger.Level >= WarnLevel {
		entry := logger.newEntry()
		entry.log(0, WarnLevel, fmt.Sprintln(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Warningln(args ...interface{}) {
	if logger.Level >= WarnLevel {
		entry := logger.newEntry()
		entry.log(0, WarnLevel, fmt.Sprintln(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Errorln(args ...interface{}) {
	if logger.Level >= ErrorLevel {
		entry := logger.newEntry()
		entry.log(0, ErrorLevel, fmt.Sprintln(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Fatalln(args ...interface{}) {
	if logger.Level >= FatalLevel {
		entry := logger.newEntry()
		entry.log(0, FatalLevel, fmt.Sprintln(args...))
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Panicln(args ...interface{}) {
	if logger.Level >= PanicLevel {
		entry := logger.newEntry()
		entry.log(0, PanicLevel, fmt.Sprintln(args...))
		logger.releaseEntry(entry)
	}
}

//When file is opened with appending mode, it's safe to
//write concurrently to a file (within 4k message on Linux).
//In these cases user can choose to disable the lock.
func (logger *Logger) SetNoLock() {
	logger.mu.Disable()
}
