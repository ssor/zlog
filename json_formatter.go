package zlog

import (
	"encoding/json"
	"fmt"
)

type JSONFormatter struct {
	// TimestampFormat sets the format used for marshaling timestamps.
	TimestampFormat string
}

func (f *JSONFormatter) Format(entry FormatterInput) ([]byte, error) {
	data := make(Fields, len(entry.GetData())+3)
	for k, v := range entry.GetData() {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/Sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}
	prefixFieldClashes(data)

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = DefaultTimestampFormat
	}

	data["time"] = entry.GetTime().Format(timestampFormat)
	data["msg"] = entry.GetMessage()
	data["level"] = entry.GetLevel().String()

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}
