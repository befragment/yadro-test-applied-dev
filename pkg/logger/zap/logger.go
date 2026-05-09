package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevel string

const (
	LogLevelError LogLevel = "error"
	LogLevelWarn  LogLevel = "warn"
	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
)

type Color string

const (
	ColorLightRed    Color = "\x1b[91m"
	ColorLightGreen  Color = "\x1b[92m"
	ColorLightYellow Color = "\x1b[93m"
	ColorLightBlue   Color = "\x1b[94m"
	ColorPurple      Color = "\x1b[95m"
	ColorCyan        Color = "\x1b[96m"
)

func parseLevel(level LogLevel) zapcore.Level {
	switch level {
	case LogLevelInfo:
		return zapcore.InfoLevel
	case LogLevelError:
		return zapcore.ErrorLevel
	case LogLevelWarn:
		return zapcore.WarnLevel
	case LogLevelDebug:
		return zapcore.DebugLevel
	default:
		return zapcore.InfoLevel
	}
}

func levelColor(l zapcore.Level) Color {
	switch l {
	case zapcore.InfoLevel:
		return ColorLightGreen
	case zapcore.DebugLevel:
		return ColorLightBlue
	case zapcore.WarnLevel:
		return ColorLightYellow
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return ColorLightRed
	default:
		return ColorLightGreen
	}
}

func levelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	const reset = "\x1b[0m"
	enc.AppendString(string(levelColor(l)) + "[" + l.CapitalString() + "]" + reset)
}

type Logger struct {
	l *zap.SugaredLogger
}

func New(level string) (*Logger, error) {
	zapLevel := parseLevel(LogLevel(level))

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        zapcore.OmitKey,
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeLevel:    levelEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	logger := zap.New(core)
	return &Logger{l: logger.Sugar()}, nil
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006/01/02 15:04:05"))
}

func (l *Logger) Debug(args ...interface{}) {
	l.l.Debug(args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.l.Debugf(format, args...)
}

func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.l.Debugw(msg, keysAndValues...)
}

func (l *Logger) Info(args ...interface{}) {
	l.l.Info(args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.l.Infof(format, args...)
}

func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.l.Infow(msg, keysAndValues...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.l.Warn(args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.l.Warnf(format, args...)
}

func (l *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	l.l.Warnw(msg, keysAndValues...)
}

func (l *Logger) Error(args ...interface{}) {
	l.l.Error(args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.l.Errorf(format, args...)
}

func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.l.Errorw(msg, keysAndValues...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.l.Fatal(args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.l.Fatalf(format, args...)
}
