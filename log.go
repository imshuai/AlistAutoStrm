package main

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
)

// Formatter is a custom logrus formatter that formats log entries.
type Formatter struct {
	Colored bool // 是否使用彩色输出
}

var (
	// 定义日志颜色代码
	// panic, fatal, error 为红色
	// warning 为黄色
	// info 为绿色
	// debug 为蓝色
	// trace 为灰色
	colors = map[string]string{
		logrus.PanicLevel.String(): "\x1b[31m", // red
		logrus.FatalLevel.String(): "\x1b[31m", // red
		logrus.ErrorLevel.String(): "\x1b[31m", // red
		logrus.WarnLevel.String():  "\x1b[33m", // yellow
		logrus.InfoLevel.String():  "\x1b[32m", // green
		logrus.DebugLevel.String(): "\x1b[34m", // blue
		logrus.TraceLevel.String(): "\x1b[30m", // gray
		"end":                      "\x1b[0m",
	}
)

// Format formats a log entry into a byte array.
//
// It takes a logrus.Entry pointer as a parameter and returns a byte array and an error.
// The byte array contains the formatted log entry.
// The error is nil if the formatting is successful.
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Initialize a bytes buffer to store the formatted log entry.
	var buffer *bytes.Buffer

	// If the entry already has a buffer assigned, use it.
	// Otherwise, create a new bytes buffer.
	if entry.Buffer != nil {
		buffer = entry.Buffer
	} else {
		buffer = &bytes.Buffer{}
	}

	// Get the timestamp of the log entry in the format "2006-01-02 15:04:05".
	timestamp := entry.Time.Format("2006-01-02 15:04:05")

	// Format the log entry into the buffer.
	// The format is "[timestamp] [log level] log message\n".
	if f.Colored {
		buffer.WriteString(fmt.Sprintf("[%s] [%s%s%s] %s\n", timestamp, colors[entry.Level.String()], entry.Level.String(), colors["end"], entry.Message))
	} else {
		buffer.WriteString(fmt.Sprintf("[%s] [%s] %s\n", timestamp, entry.Level.String(), entry.Message))
	}
	// Return the formatted log entry as a byte array and a nil error.
	return buffer.Bytes(), nil
}
