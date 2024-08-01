package logger

import (
	"alert_gateway/config"
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

	// 定义日志输出模式
	var logOutput io.Writer

	logPath, ok := logConfig["path"].(string)
	if !ok {
		logPath = ""
	}

	logType, ok := logConfig["type"].(string)
	if !ok {
		logType = "console"
	}

	switch logType {
	case "file":
		if logPath == "" {
			log.Fatalf("当日志类型是file时日志的路径必须设置")
		}
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("打开日志文件失败: %v", err)
		}
		// 设置日志输出模式为文本
		logOutput = file
	case "all":
		if logPath == "" {
			log.Fatalf("当日志类型为all时日志路径必须设置")
		}
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("打开日志文件失败: %v", err)
		}
		// 设置日志输出模式为文本和终端
		logOutput = io.MultiWriter(os.Stdout, file)
	case "console":
		fallthrough
	default:
		//设置日志输出模式为终端
		logOutput = os.Stdout
	}

	// 读取配置中指定的日志级别
	logLevel, ok := logConfig["level"].(string)
	if !ok {
		logLevel = "info"
	}

	// 为不同级别的日志定义格式
	InfoLogger = log.New(logOutput, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(logOutput, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(logOutput, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

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
