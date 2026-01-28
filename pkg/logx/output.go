package logx

type Logger interface {
	Log(entry Entry)
}

func SetOutput(logger Logger) {
	if logger != nil {
		defaultLogger.output = logger
	}
}

func SetTraceIDFunc(fn TraceIDFunc) {
	if fn != nil {
		defaultLogger.traceIDFunc = fn
	}
}
