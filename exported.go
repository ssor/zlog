package zlog

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

var (
	// std is the name of the standard logger in stdlib `log`
	std                      *Logger
	exportedDefaultCallDepth = 6
	loggers                  = []*Logger{}
)

func StandardLogger() *Logger {
	if std == nil {
		std = New()
		loggers = append(loggers, std)
	}
	return std
}

// SetOutput sets the standard logger output.
func SetOutput(out io.Writer) {
	for _, logger := range loggers {
		logger.SetOutput(out)
	}
}

func SetPrintLineNumber(b bool) {
	if b {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}
}

// SetLevel sets the standard logger level.
func SetLevel(level Level) {
	logger := StandardLogger()
	logger.SetLevel(level)
}

// GetLevel returns the standard logger level.
func GetLevel() Level {
	logger := StandardLogger()
	return logger.Level
}

// WithError creates an entry from the standard logger and adds an error to it, using the value defined in ErrorKey as key.
func WithError(err error) *Entry {
	logger := StandardLogger()
	return logger.WithField(ErrorKey, err)
}

func DumpStacks() {
	dumpStacks()
}

// WithField creates an entry from the standard logger and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithField(key string, value interface{}) *Entry {
	logger := StandardLogger()
	return logger.WithField(key, value)
}

func WithMultiLines(key, longStr string) *Entry {
	logger := StandardLogger()

	var entry *Entry
	lns := strings.Split(longStr, "\n")
	for index, ln := range lns {
		if len(ln) <= 0 {
			continue
		}
		entry = logger.WithField(fmt.Sprintf("%s-%d", key, index), ln)
	}
	return entry
}

func WithLongString(key, longStr, sep string) *Entry {
	logger := StandardLogger()

	var entry *Entry
	lns := strings.Split(longStr, sep)
	for index, ln := range lns {
		if len(ln) <= 0 {
			continue
		}

		entry = logger.WithField(fmt.Sprintf("%s-%d", key, index), ln)
	}
	return entry
}

func WithStruct(value interface{}) *Entry {
	bs, err := json.Marshal(value)
	if err != nil {
		return WithError(err)
	}
	return WithJsonRaw(bs)
}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithFields(fields Fields) *Entry {
	logger := StandardLogger()

	return logger.WithFields(fields)
}

func WithJsonRaw(bs []byte) *Entry {
	logger := StandardLogger()

	return logger.WithJsonRaw(bs)
}

func AddFields(args ...interface{}) *Entry {
	if len(args) <= 0 {
		return WithFields(map[string]interface{}{})
	}

	if len(args)%2 != 0 {
		args = append(args, "")
	}
	max := len(args) - 2
	data := make(map[string]interface{})
	for i := 0; i <= max; i = i + 2 {
		data[args[i].(string)] = args[i+1]
	}
	return WithFields(data)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	logger := StandardLogger()

	logger.Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	logger := StandardLogger()

	logger.Print(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	logger := StandardLogger()

	logger.Info(args...)
}

func Pass(args ...interface{}) {
	logger := StandardLogger()

	logger.Pass(args...)
}

func Passf(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.Passf(format, args...)
}

func Failed(args ...interface{}) {
	logger := StandardLogger()

	logger.Failed(args...)
}

func Failedf(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.Failedf(format, args...)
}

func Success(args ...interface{}) {
	logger := StandardLogger()

	logger.Success(args...)
}

func Successf(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.Successf(format, args...)
}

func Highlight(args ...interface{}) {
	logger := StandardLogger()

	logger.highlight(exportedDefaultCallDepth, args...)
}

func Highlightf(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.highlight(exportedDefaultCallDepth, fmt.Sprintf(format, args...))
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	logger := StandardLogger()

	logger.Warn(args...)
	//WithFields(nil).Warn(args)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	logger := StandardLogger()

	logger.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	logger := StandardLogger()

	logger.Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	logger := StandardLogger()

	logger.Fatal(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.Debugf(format, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.Warnf(format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func Warningf(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.Warningf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	logger := StandardLogger()

	logger.Fatalf(format, args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...interface{}) {
	logger := StandardLogger()

	logger.Debugln(args...)
}

// Println logs a message at level Info on the standard logger.
func Println(args ...interface{}) {
	logger := StandardLogger()

	logger.Println(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(args ...interface{}) {
	logger := StandardLogger()

	logger.Infoln(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(args ...interface{}) {
	logger := StandardLogger()

	logger.Warnln(args...)
}

// Warningln logs a message at level Warn on the standard logger.
func Warningln(args ...interface{}) {
	logger := StandardLogger()

	logger.Warningln(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(args ...interface{}) {
	logger := StandardLogger()

	logger.Errorln(args...)
}

// Panicln logs a message at level Panic on the standard logger.
func Panicln(args ...interface{}) {
	logger := StandardLogger()

	logger.Panicln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func Fatalln(args ...interface{}) {
	logger := StandardLogger()

	logger.Fatalln(args...)
}

func PrettyJson(j []byte) {
	logger := StandardLogger()

	logger.Println(prettyJSON(j))
}
