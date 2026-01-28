package logx

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

type Level uint8

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

type Fields map[string]interface{}

type Entry struct {
	Time      time.Time
	Level     Level
	Message   string
	Caller    string
	RelatedID string
	TraceID   string
	Fields    Fields
}

type LoggerWrapper struct {
	traceIDFunc TraceIDFunc // 如何拿到 TraceID
	output      Logger      // 如何处理报错信息
}

var defaultLogger = &LoggerWrapper{
	traceIDFunc: defaultTraceIDFunc,
	output:      NewConsoleLogger(),
}

func (l *LoggerWrapper) logEntry(
	ctx context.Context,
	level Level,
	relatedID,
	msg string,
	fields ...Fields,
) {
	_, file, line, ok := runtime.Caller(3)
	caller := "unknown:0"
	if ok {
		caller = fmt.Sprintf("%s:%d", file, line)
	}

	entry := Entry{
		Time:      time.Now(),
		Level:     level,
		Message:   msg,
		Caller:    caller,
		RelatedID: relatedID,
		TraceID:   l.traceIDFunc(ctx),
		Fields:    mergeFields(fields...),
	}

	l.output.Log(entry)
}

/*
日志分为 Info, Error, DEBUG
目前只考虑 Info 和 Error
如果系统是按预期正常运行的，就打 Info
否则打 Error

日志解决的是“事后”的问题
*/

func Info(ctx context.Context, relatedID string, msg ...any) {
	defaultLogger.logEntry(ctx, LevelInfo, relatedID, fmt.Sprint(msg...))
}

func Infos(ctx context.Context, relatedID, msg string, fields ...Fields) {
	defaultLogger.logEntry(ctx, LevelInfo, relatedID, msg, fields...)
}

func Warn(ctx context.Context, relatedID string, msg ...any) {
	defaultLogger.logEntry(ctx, LevelWarn, relatedID, fmt.Sprint(msg...))
}

func Error(ctx context.Context, relatedID string, msg ...any) {
	defaultLogger.logEntry(ctx, LevelError, relatedID, fmt.Sprint(msg...))
}

func Errors(ctx context.Context, relatedID, msg string, fields ...Fields) {
	defaultLogger.logEntry(ctx, LevelError, relatedID, msg, fields...)
}

// 启动/任务场景（无 ctx，不推荐频繁用）
func Infow(relatedID string, msg ...any) {
	defaultLogger.logEntry(nil, LevelInfo, relatedID, fmt.Sprint(msg...))
}
