package main

import (
	"bytes"
	"fmt"
	"sync/atomic"

	"github.com/sirupsen/logrus"
	"github.com/vbauerster/mpb/v8"
)

type StatLogger struct {
	*logrus.Logger // 内部使用logrus库
	count          *atomic.Int64
	bar            *mpb.Bar
}

func NewLogger() *StatLogger {
	l := &StatLogger{
		logrus.New(),
		&atomic.Int64{},
		&mpb.Bar{},
	}
	l.count.Store(0)
	return l
}

func (l *StatLogger) SetBar(bar *mpb.Bar) {
	l.bar = bar
}

func (l *StatLogger) Increment() {
	l.bar.Increment()
}

func (l *StatLogger) SetTotal(total int64) {
	l.bar.SetTotal(total, false)
}

func (l *StatLogger) SetCurrent(current int64) {
	l.bar.SetCurrent(current)
}

func (l *StatLogger) GetCurrent() int64 {
	return l.bar.Current()
}

func (l *StatLogger) Add(n int64) {
	l.count.Add(n)
}

func (l *StatLogger) GetCount() int64 {
	return l.count.Load()
}

// func (l *Logger) SetFormatter(f logrus.Formatter) {
// 	l.l.SetFormatter(f)
// }

// func (l *Logger) Trace(args ...interface{}) {
//     l.l.Trace(args...)
// }

// func (l *Logger) Tracef(format string, args ...interface{}) {
//     l.l.Tracef(format, args...)
// }

// func (l *Logger) Debug(args ...interface{}) {
//     l.l.Debug(args...)
// }

// func (l *Logger) Debugf(format string, args ...interface{}) {
//     l.l.Debugf(format, args...)
// }

// func (l *Logger) Info(args ...interface{}) {
//     l.l.Info(args...)
// }

// func (l *Logger) Infof(format string, args ...interface{}) {
//     l.l.Infof(format, args...)
// }

// func (l *Logger) Warn(args ...interface{}) {
//     l.l.Warn(args...)
// }

// func (l *Logger) Warnf(format string, args ...interface{}) {
//     l.l.Warnf(format, args...)
// }

// func (l *Logger) Error(args ...interface{}) {
//     l.l.Error(args...)
// }

// func (l *Logger) Errorf(format string, args ...interface{}) {
//     l.l.Errorf(format, args...)
// }

// func (l *Logger) Fatal(args ...interface{}) {
//     l.l.Fatal(args...)
// }

// func (l *Logger) Fatalf(format string, args ...interface{}) {
//     l.l.Fatalf(format, args...)
// }

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
