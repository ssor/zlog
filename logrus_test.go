package zlog

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func LogAndAssertText(t *testing.T, log func(*Logger), assertions func(fields map[string]string)) {
	var buffer bytes.Buffer

	logger := New()
	logger.Out = &buffer
	logger.Formatter = &TextFormatter{
		DisableColors: true,
	}

	log(logger)

	fields := make(map[string]string)
	for _, kv := range strings.Split(buffer.String(), " ") {
		if !strings.Contains(kv, "=") {
			continue
		}
		kvArr := strings.Split(kv, "=")
		key := strings.TrimSpace(kvArr[0])
		val := kvArr[1]
		if kvArr[1][0] == '"' {
			var err error
			val, err = strconv.Unquote(val)
			assert.NoError(t, err)
		}
		fields[key] = val
	}
	assertions(fields)
}


func TestWithFieldsShouldAllowAssignments(t *testing.T) {
	var buffer bytes.Buffer
	var fields Fields

	logger := New()
	logger.Out = &buffer
	logger.Formatter = new(TextFormatter)

	localLog := logger.WithFields(Fields{
		"key1": "value1",
	})

	localLog.WithField("key2", "value2").Info("test")
	err := json.Unmarshal(buffer.Bytes(), &fields)
	assert.Nil(t, err)

	assert.Equal(t, "value2", fields["key2"])
	assert.Equal(t, "value1", fields["key1"])

	buffer = bytes.Buffer{}
	fields = Fields{}
	localLog.Info("test")
	err = json.Unmarshal(buffer.Bytes(), &fields)
	assert.Nil(t, err)

	_, ok := fields["key2"]
	assert.Equal(t, false, ok)
	assert.Equal(t, "value1", fields["key1"])
}

func TestDefaultFieldsAreNotPrefixed(t *testing.T) {
	LogAndAssertText(t, func(log *Logger) {
		ll := log.WithField("herp", "derp")
		ll.Info("hello")
		ll.Info("bye")
	}, func(fields map[string]string) {
		for _, fieldName := range []string{"fields.level", "fields.time", "fields.msg"} {
			if _, ok := fields[fieldName]; ok {
				t.Fatalf("should not have prefixed %q: %v", fieldName, fields)
			}
		}
	})
}

func TestDoubleLoggingDoesntPrefixPreviousFields(t *testing.T) {

	var buffer bytes.Buffer
	var fields Fields

	logger := New()
	logger.Out = &buffer
	logger.Formatter = new(TextFormatter)

	llog := logger.WithField("context", "eating raw fish")

	llog.Info("looks delicious")

	err := json.Unmarshal(buffer.Bytes(), &fields)
	assert.NoError(t, err, "should have decoded first message")
	assert.Equal(t, len(fields), 4, "should only have msg/time/level/context fields")
	assert.Equal(t, fields["msg"], "looks delicious")
	assert.Equal(t, fields["context"], "eating raw fish")

	buffer.Reset()

	llog.Warn("omg it is!")

	err = json.Unmarshal(buffer.Bytes(), &fields)
	assert.NoError(t, err, "should have decoded second message")
	assert.Equal(t, len(fields), 4, "should only have msg/time/level/context fields")
	assert.Equal(t, fields["msg"], "omg it is!")
	assert.Equal(t, fields["context"], "eating raw fish")
	assert.Nil(t, fields["fields.msg"], "should not have prefixed previous `msg` entry")

}

func TestConvertLevelToString(t *testing.T) {
	assert.Equal(t, "debug", DebugLevel.String())
	assert.Equal(t, "info", InfoLevel.String())
	assert.Equal(t, "warning", WarnLevel.String())
	assert.Equal(t, "error", ErrorLevel.String())
	assert.Equal(t, "fatal", FatalLevel.String())
	assert.Equal(t, "panic", PanicLevel.String())
}

func TestParseLevel(t *testing.T) {
	l, err := ParseLevel("panic")
	assert.Nil(t, err)
	assert.Equal(t, PanicLevel, l)

	l, err = ParseLevel("PANIC")
	assert.Nil(t, err)
	assert.Equal(t, PanicLevel, l)

	l, err = ParseLevel("fatal")
	assert.Nil(t, err)
	assert.Equal(t, FatalLevel, l)

	l, err = ParseLevel("FATAL")
	assert.Nil(t, err)
	assert.Equal(t, FatalLevel, l)

	l, err = ParseLevel("error")
	assert.Nil(t, err)
	assert.Equal(t, ErrorLevel, l)

	l, err = ParseLevel("ERROR")
	assert.Nil(t, err)
	assert.Equal(t, ErrorLevel, l)

	l, err = ParseLevel("warn")
	assert.Nil(t, err)
	assert.Equal(t, WarnLevel, l)

	l, err = ParseLevel("WARN")
	assert.Nil(t, err)
	assert.Equal(t, WarnLevel, l)

	l, err = ParseLevel("warning")
	assert.Nil(t, err)
	assert.Equal(t, WarnLevel, l)

	l, err = ParseLevel("WARNING")
	assert.Nil(t, err)
	assert.Equal(t, WarnLevel, l)

	l, err = ParseLevel("info")
	assert.Nil(t, err)
	assert.Equal(t, InfoLevel, l)

	l, err = ParseLevel("INFO")
	assert.Nil(t, err)
	assert.Equal(t, InfoLevel, l)

	l, err = ParseLevel("debug")
	assert.Nil(t, err)
	assert.Equal(t, DebugLevel, l)

	l, err = ParseLevel("DEBUG")
	assert.Nil(t, err)
	assert.Equal(t, DebugLevel, l)

	l, err = ParseLevel("invalid")
	assert.Equal(t, "not a valid logrus Level: \"invalid\"", err.Error())
}

func TestGetSetLevelRace(t *testing.T) {
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				SetLevel(InfoLevel)
			} else {
				GetLevel()
			}
		}(i)

	}
	wg.Wait()
}

func TestLoggingRace(t *testing.T) {
	logger := New()

	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			logger.Info("info")
			wg.Done()
		}()
	}
	wg.Wait()
}
