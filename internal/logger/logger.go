package logger

import (
	"context"
	"os"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/constants/contextkeys"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.SugaredLogger

func Initialize(level string, format string) error {
	cfg := createLogConfig(format)
	cfg.Level = parseLogLevel(level)
	configureEncoder(&cfg)
	cfg.OutputPaths = []string{constants.LogOutputStdout}
	cfg.ErrorOutputPaths = []string{constants.LogOutputStderr}

	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}
	L = logger.Sugar()
	return nil
}

func createLogConfig(format string) zap.Config {
	if format == constants.LogFormatJSON {
		return zap.NewProductionConfig()
	}
	return zap.NewDevelopmentConfig()
}

func parseLogLevel(level string) zap.AtomicLevel {
	switch level {
	case constants.LogLevelDebug:
		return zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case constants.LogLevelWarn:
		return zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case constants.LogLevelError:
		return zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
}

func configureEncoder(cfg *zap.Config) {
	cfg.EncoderConfig.TimeKey = constants.LogEncoderTimeKey
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.MessageKey = constants.LogEncoderMessageKey
	cfg.EncoderConfig.LevelKey = constants.LogEncoderLevelKey
	cfg.EncoderConfig.CallerKey = constants.LogEncoderCallerKey
}

func Ctx(ctx context.Context) *zap.SugaredLogger {
	if L == nil {
		l, _ := zap.NewProduction()
		L = l.Sugar()
	}
	logger := L
	if rid, ok := ctx.Value(contextkeys.RequestID).(string); ok {
		logger = logger.With(constants.LogKeyRequestID, rid)
	}
	if tid, ok := ctx.Value(contextkeys.TraceID).(string); ok && tid != "" {
		logger = logger.With(constants.LogFieldTraceID, tid)
	}
	if sid, ok := ctx.Value(contextkeys.SessionID).(string); ok && sid != "" {
		logger = logger.With(constants.LogFieldSessionID, sid)
	}
	return logger
}

func Info(msg string, keysAndValues ...interface{})  { if L != nil { L.Infow(msg, keysAndValues...) } }
func Debug(msg string, keysAndValues ...interface{}) { if L != nil { L.Debugw(msg, keysAndValues...) } }
func Warn(msg string, keysAndValues ...interface{})  { if L != nil { L.Warnw(msg, keysAndValues...) } }
func Error(msg string, keysAndValues ...interface{}) { if L != nil { L.Errorw(msg, keysAndValues...) } }
func Fatal(msg string, keysAndValues ...interface{}) {
	if L == nil { os.Exit(1) }
	L.Fatalw(msg, keysAndValues...)
}
func Sync() { if L != nil { _ = L.Sync() } }
