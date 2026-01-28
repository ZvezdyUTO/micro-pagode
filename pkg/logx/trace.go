package logx

import "context"

type TraceIDFunc func(context.Context) string

type traceIDKey struct{}

func SetTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if traceID == "" {
		traceID = generateTraceID()
	}
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(traceIDKey{}).(string); ok {
		return v
	}
	return ""
}

func defaultTraceIDFunc(ctx context.Context) string {
	if traceID := GetTraceID(ctx); traceID != "" {
		return traceID
	}
	return generateTraceID()
}
