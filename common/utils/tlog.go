package utils

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"time"
)

var TLog *zap.Logger

func InitLog(isdebug bool, maxday int, logfilename string, lv string) {
	//初始化encoder配置
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		// TODO: consider using a byte-oriented API to save an allocation.
		enc.AppendString("[" + caller.TrimmedPath() + "]")
	}
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeDuration = func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendInt64(int64(d) / 1000000)
	}
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.MessageKey = "msgkey"
	encoderCfg.LevelKey = "lvkey"
	encoderCfg.TimeKey = "timekey"
	encoderCfg.CallerKey = "callerkey"

	/*
		idx := strings.Index(logfilename, ".log")
		if idx == -1 {
			logfilename = logfilename + "Ex"
		} else {
			logfilename = logfilename[0:idx] + "Ex" + logfilename[idx:]
		}
	*/
	infoWriter := getWriter(logfilename, maxday)
	w := zapcore.AddSync(infoWriter)

	var level zapcore.Level
	switch lv {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	var core zapcore.Core
	//调试模式
	if isdebug {

		consoleErrors := zapcore.Lock(os.Stderr)

		core = zapcore.NewTee(
			zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), consoleErrors, zap.DebugLevel),
			zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), w, zap.DebugLevel),
		)
	} else {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			w,
			level,
		)
	}

	TLog = zap.New(core, zap.AddCaller())
	// defer TLog.Sync() //main函数退出时使用
}

func getWriter(filename string, maxDays int) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名.log.YYmmddHH
	// 保存maxDays天内的日志，每1小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		filename+".%Y%m%d%H",
		//rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*time.Duration(24*maxDays)),
		rotatelogs.WithRotationTime(time.Hour),
	)

	if err != nil {
		panic(err)
	}
	return hook
}
