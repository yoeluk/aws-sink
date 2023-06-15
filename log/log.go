package log

import (
	"encoding/json"
	"fmt"
	"time"
)

type LogEvent struct {
	Level string `json:"level"`
	Msg   string `json:"msg"`
	Time  string `json:"time"`
}

func Info(msg string) {
	newLogEvent("info", msg).print()
}

func Debug(msg string) {
	newLogEvent("debug", msg).print()
}

func Warn(msg string) {
	newLogEvent("warn", msg).print()
}

func Error(msg string) {
	newLogEvent("error", msg).print()
}

func newLogEvent(level, msg string) *LogEvent {
	return &LogEvent{
		Level: level,
		Msg:   msg,
	}
}

func (logEvent *LogEvent) print() {
	t := time.Now()
	logEvent.Time = t.UTC().Format(time.RFC3339)
	jsonLogEvent, _ := json.Marshal(*logEvent)
	fmt.Printf("%s %s\n", t.Format(time.RFC3339), string(jsonLogEvent))
}
