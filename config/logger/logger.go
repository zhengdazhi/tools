package logger

import (
	"config/config"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

func InitLogger(cfg *config.Config) {
	logConfig := cfg.Log

	// 设置日志输出
	var logOutput io.Writer

	// 根据配置文件设置日志文件路径
	logPath, ok := logConfig["path"].(string)
	if !ok {
		logPath = ""
	}

	// 根据配置文件设置日志输出类型
	logType, ok := logConfig["type"].(string)
	if !ok {
		logType = "console"
	}

	switch logType {
	case "file":
		if logPath == "" {
			log.Fatalf("Log path must be specified when log type is 'file'")
		}
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		logOutput = file
	case "all":
		if logPath == "" {
			log.Fatalf("Log path must be specified when log type is 'all'")
		}
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		logOutput = io.MultiWriter(os.Stdout, file)
	case "console":
		fallthrough
	default:
		logOutput = os.Stdout
	}

	// 设置日志级别
	logLevel, ok := logConfig["level"].(string)
	if !ok {
		logLevel = "info"
	}

	InfoLogger = log.New(logOutput, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(logOutput, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(io.Discard, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile) // 默认丢弃

	// 判断获取的日志级别
	switch logLevel {
	case "debug":
		DebugLogger.SetOutput(logOutput)
	case "info":
		DebugLogger.SetOutput(io.Discard)
	case "error":
		InfoLogger.SetOutput(io.Discard)
		DebugLogger.SetOutput(io.Discard)
	}
}

func Info(v ...interface{}) {
	InfoLogger.Println(v...)
}

func Infof(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

func Error(v ...interface{}) {
	ErrorLogger.Println(v...)
}

func Errorf(format string, v ...interface{}) {
	ErrorLogger.Printf(format, v...)
}

func Debug(v ...interface{}) {
	//DebugLogger.Println(v...)
	DebugLogger.Output(2, fmt.Sprintf("%s: %s", getCallerInfo(), fmt.Sprintln(v...)))
}

func Debugf(format string, v ...interface{}) {
	//DebugLogger.Printf(format, v...)
	DebugLogger.Output(2, fmt.Sprintf("%s: %s", getCallerInfo(), fmt.Sprintf(format, v...)))
}

func getCallerInfo() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	return fn.Name()
}
