package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"path/filepath"
	"strings"
)

var logger *zap.SugaredLogger

var levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

func StartLogger(logName string, logLevel string) {
	appRootPath := "/var/log"
	fileName := strings.Join([]string{appRootPath, "mysql-agent", logName}, string(filepath.Separator))
	level := getLoggerLevel(logLevel)
	syncWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:  fileName,
		MaxSize:   100,
		MaxAge:    7,
		LocalTime: true,
		Compress:  true,
	})
	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), syncWriter, zap.NewAtomicLevelAt(level))
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	logger = zapLogger.Sugar()
}

func Info(msg string, args ...interface{}) {
	logger.Infof(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	logger.Warnf(msg, args...)
}

func Error(msg string, args ...interface{}) {
	logger.Errorf(msg, args...)
}

func Fatal(msg string, args ...interface{}) {
	logger.Fatalf(msg, args...)
}

func Panic(msg string, args ...interface{}) {
	logger.Panicf(msg, args...)
}

func DPanic(msg string, args ...interface{}) {
	logger.DPanicf(msg, args...)
}
