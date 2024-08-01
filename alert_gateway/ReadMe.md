## 创建项目结构

.
├── app.log
├── apps
│   └── web.go
├── cmd
│   └── main.go
├── config
│   └── config.go
├── config.toml
├── go.mod
├── go.sum
├── logger
│   └── logger.go
└── test
    └── test.json

### 入口文件

cmd/main.go

```go
package main

import (
	"alert_gateway/apps"
	"alert_gateway/config"
	"alert_gateway/logger"
	"flag"
	"log"
)

func main() {
	var help bool
	var configPath string

	flag.BoolVar(&help, "help", false, "show help informaction")
	flag.StringVar(&configPath, "config", "config.toml", "path to config file")

	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.InitLogger(cfg)

	switch {
	case help:
		flag.PrintDefaults()
	default:
		// 启动应用
		apps.Run(cfg)
	}
}

```

配置加载

config/config.go

```go
package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	App map[string]any
	Log map[string]any
}

func NewConfig() *Config {
	return &Config{
		App: make(map[string]any),
		Log: make(map[string]any),
	}
}

func LoadConfig(configPath string) (*Config, error) {
	config := NewConfig()
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("error loading config file file :%w", err)
	}
	return config, nil
}
```

### 日志模块

logger/logger.go

```go
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

```

### 应用模块

apps/web.go

```go
package apps

import (
	"alert_gateway/config"
	"alert_gateway/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func Run(cfg *config.Config) {
	logger.Info("info log")
	logger.Error("error log")
	logger.Debug("debug log")

	appConfig := cfg.App
	addr := fmt.Sprintf("%s:%v", appConfig["listen"], appConfig["port"])
	logger.Infof("Server running on %s", addr)

	server := http.Server{
		Addr: addr,
	}
	http.HandleFunc("/", index(cfg))
	if err := server.ListenAndServe(); err != nil {
		logger.Errorf("Server error: %v", err)
	}
}

type IndexData struct {
	Title string `json:"tile"`
	Desc  string `json:"desc"`
}

// 报警结构体中的Alerts数组内的报警内容
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// 定义alertmanager发送的报警结构体
type AlertData struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
}

type MarkdownMessage struct {
	MsgType  string   `json:"msgtype"`
	Markdown Markdown `json:"markdown"`
}

type TextMessage struct {
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
}

type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Theme string `json:"theme"`
}

type Text struct {
	Content string `json:"content"`
}

// func index(w http.ResponseWriter, r *http.Request) {
//  logger.Debugf("Received request at %s", r.URL.Path)

//  switch r.Method {
//  case http.MethodGet:
//      handleGet(w, r)
//  case http.MethodPost:
//      handlePost(w, r)
//  default:
//      http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//  }
// }

func index(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debugf("Received request at %s", r.URL.Path)

		switch r.Method {
		case http.MethodGet:
			handleGet(w, r)
		case http.MethodPost:
			handlePost(w, r, cfg)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// 处理get请求
func handleGet(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Handling GET request")
	w.Header().Set("Content-Type", "application/json")

	helpMessage := `使用以下命令发送POST请求: curl -X POST http://localhost:8080/ -H "Content-Type: application/json" -d '{"receiver": "web\\.hook","status": "firing","alerts": [{"status": "firing","labels": {"alertname": "主机状态","alias": "temperature","instance": "10.10.1.121:8000","job": "cpu temperature","severity": "critical"},"annotations": {"description": "测试环境 10.10.1.121:8000:服务器关闭","summary": "Instance 10.10.1.121:8000:服务器关闭"},"startsAt": "2024-07-29T04:25:09.673Z","endsAt": "0001-01-01T00:00:00Z","generatorURL": "http://localhost.localdomain:9090/graph?g0.expr=up+%3D%3D+0\u0026g0.tab=1","fingerprint": "56c8b0d55e0b6050"},{"status": "resolved","labels": {"alertname": "主机状态","alias": "mysql","instance": "10.10.1.36:9104","job": "db","severity": "critical"},"annotations": {"description": "测试环境 10.10.1.36:9104:服务器关闭","summary": "Instance 10.10.1.36:9104:服务器关闭"},"startsAt": "2024-07-30T14:18:54.673Z","endsAt": "2024-07-30T14:19:24.673Z","generatorURL": "http://localhost.localdomain:9090/graph?g0.expr=up+%3D%3D+0\u0026g0.tab=1","fingerprint": "aab48de6a92cc407"},{"status": "resolved","labels": {"alertname": "主机状态","alias": "openstack","instance": "10.10.2.236:9100","job": "openstack","severity": "critical"},"annotations": {"description": "测试环境 10.10.2.236:9100:服务器关闭","summary": "Instance 10.10.2.236:9100:服务器关闭"},"startsAt": "2024-07-29T15:48:09.673Z","endsAt": "2024-07-29T15:48:24.673Z","generatorURL": "http://localhost.localdomain:9090/graph?g0.expr=up+%3D%3D+0\u0026g0.tab=1","fingerprint": "4c02eb2cd7a86ce8"},{"status": "resolved","labels": {"alertname": "cpu使用率过高","instance": "10.10.2.235:9100","severity": "warning"},"annotations": {"description": "测试环境 10.10.2.235:9100 of job cpu使用率超过80%,当前使用率[72.004166667078].","summary": "Instance 10.10.2.235:9100 cpu使用率过高"},"startsAt": "2024-07-30T20:19:09.673Z","endsAt": "2024-07-30T20:22:54.673Z","generatorURL": "http://localhost.localdomain:9090/graph?g0.expr=100+-+avg+by%28instance%29+%28irate%28node_cpu_seconds_total%7Bmode%3D%22idle%22%7D%5B5m%5D%29%29+%2A+100+%3E+60\u0026g0.tab=1","fingerprint": "3b5ccd9658ce3997"}],"groupLabels": {},"commonLabels": {},"commonAnnotations": {},"externalURL": "http://localhost.localdomain:9093","version": "4","groupKey": "{}:{}","truncatedAlerts": 0}'`
	helpMessage = strings.ReplaceAll(helpMessage, "\n", "")
	helpMessage = strings.ReplaceAll(helpMessage, "\r", "")
	jsonStr, _ := json.Marshal(map[string]string{"help": helpMessage})
	w.Write(jsonStr)
}

// 处理post请求
func handlePost(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	logger.Debug("Handling POST request")
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		logger.Errorf("Failed to read request body: %v", err)
		return
	}
	// 调试打印请求
	// logger.Debugf("----- %s ------", getTime())
	// logger.Debugf("Received body: %s \n", string(body))
	// logger.Debugf("----- %s ------", getTime())

	var alertData AlertData
	err = json.Unmarshal(body, &alertData)
	if err != nil {
		http.Error(w, "解析json数据失败", http.StatusBadRequest)
		logger.Errorf("Failed to parse JSON data: %v", err)
	}

	// 定义一个map类型的切片，存放发送钉钉消息的结果数据
	var responses []map[string]interface{}

	// 读取配置文件设置发现钉钉的消息类型
	messageType := "markdown" // 默认消息类型
	if cfgMessageType, ok := cfg.App["messageType"].(string); ok {
		messageType = cfgMessageType
	}

	// 循环发送报警信息到钉钉
	for _, alert := range alertData.Alerts {
		var message string
		if alert.Status == "resolved" {
			if messageType == "text" {
				message = createText("恢复", alert)
			} else {
				message = createMarkDown("恢复", "#00FF00", alert)
			}
		} else {
			if messageType == "text" {
				message = createText("故障", alert)
			} else {
				message = createMarkDown("故障", "#FF0000", alert)
			}
		}

		respMsg, err := sendMsg(message, messageType)
		response := map[string]interface{}{
			"alert":   alert.Labels["instance"],
			"respMsg": respMsg,
			"error":   err,
		}

		responses = append(responses, response)

		if err != nil {
			logger.Errorf("Failed to send message: %v", err)
		} else {
			logger.Debugf("Response message: %s", respMsg)
		}
	}

	responseBody, err := json.Marshal(responses)
	if err != nil {
		http.Error(w, "Failed to marshal response JSON", http.StatusInternalServerError)
		logger.Errorf("Failed to marshal response JSON: %v", err)
		return
	}

	// Echo the response back to the client
	w.Write(responseBody)
}

func getTime() string {
	// now := time.Now()
	// return fmt.Sprintf(now.Format("2006-01-02 15:04:05"))
	return time.Now().Format("2006-01-02 15:04:05")
}

func timeFormat(timeStr string) (string, error) {
	// 待转换的时间字符串
	//timeStr := "2024-07-30T20:19:09.673Z"

	// 解析时间字符串，指定输入格式
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		fmt.Println("解析时间错误:", err)
		return "", err
	}

	// 格式化为目标格式
	newTimeStr := t.Format("2006-01-02 15:04:05")
	return newTimeStr, nil
}

func createMarkDown(title, color string, alert Alert) string {
	var markdown bytes.Buffer

	// Construct the Markdown string
	markdown.WriteString(fmt.Sprintf("# <font color=%s>%s</font>\n\n", color, title))
	//markdown.WriteString("## Items\n\n")
	markdown.WriteString(fmt.Sprintf("- summary: %s \n", alert.Annotations["summary"]))
	for key, item := range alert.Labels {
		markdown.WriteString(fmt.Sprintf("- %s: %s\n", key, item))
	}
	startsAt, err := timeFormat(alert.StartsAt)
	if err != nil {
		logger.Error("create Markdown err")
	}
	markdown.WriteString(fmt.Sprintf("- StartsAt: %s \n", startsAt))

	return markdown.String()
}

func createText(title string, alert Alert) string {
	var textContent bytes.Buffer
	textContent.WriteString(fmt.Sprintf("content: %s\n", title))
	textContent.WriteString(fmt.Sprintf("summary: %s \n", alert.Annotations["summary"]))
	for key, item := range alert.Labels {
		textContent.WriteString(fmt.Sprintf("%s: %s\n", key, item))
	}
	startsAt, err := timeFormat(alert.StartsAt)
	if err != nil {
		logger.Error("create text err")
	}
	textContent.WriteString(fmt.Sprintf("StartsAt: %s \n", startsAt))

	return textContent.String()
}

func sendMsg(message, messageType string) (string, error) {
	//secret := "SEC914ce8b70b3f05caa6b221d8da4d58b886bcd8baea8b51bd0a0e163460313a9b"
	token := "37224eaeafda63f1d98a7daca8b2b3f591d24e713f6109053678d94109482f01"
	url := "https://oapi.dingtalk.com/robot/send?access_token=" + token

	var sendDataBytes []byte
	var err error
	if messageType == "text" {
		sendData := TextMessage{
			MsgType: "text",
			Text: Text{
				Content: message,
			},
		}
		sendDataBytes, err = json.Marshal(sendData)
	} else {
		sendData := MarkdownMessage{
			MsgType: "markdown",
			Markdown: Markdown{
				Title: "消息",
				Text:  message,
				Theme: "white",
			},
		}
		sendDataBytes, err = json.Marshal(sendData)
	}

	if err != nil {
		logger.Errorf("Failed to marshal JSON: %v", err)
		return "", err
	}
	logger.Debug(string(sendDataBytes))

	reqBody := bytes.NewBuffer(sendDataBytes)

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
		return "", err
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Failed to send request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// 获取发送后的相应
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Failed to read response body: %v", err)
		return "", err
	}
	// 处理响应
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Message sent successfully")
		//fmt.Println(string(respBody))
		logger.Debugf("发送钉钉相应 %s", string(respBody))
		return string(respBody), nil
	} else {
		fmt.Printf("Failed to send message, status code: %d\n", resp.StatusCode)
		//fmt.Println(string(respBody))
		logger.Debugf("发送钉钉相应 %s", string(respBody))
		return string(respBody), nil
	}
}

```

### 配置文件

config.toml

```toml
[app]
port = "5000"
listen = "127.0.0.1"
token = "91ff7ab2cadb126093dde158f702466384e7810151d8c3e4b60a37a7bc8bc799"
secret = "SECc11fd69753067a579058b3c6d012d3d24d01ad50592280c4e11cb0ea6872c6eb"
messageType = "markdown"  # "markdown" "text"

[log]
type = "all"
path = "app.log"
level = "debug"

```

## 优化代码

### web/utils.go

将独立的函数放到单独文件

```go
package apps

import (
	"alert_gateway/logger"
	"bytes"
	"fmt"
	"time"
)

func getTime() string {
	// now := time.Now()
	// return fmt.Sprintf(now.Format("2006-01-02 15:04:05"))
	return time.Now().Format("2006-01-02 15:04:05")
}

func timeFormat(timeStr string) (string, error) {
	// 待转换的时间字符串
	//timeStr := "2024-07-30T20:19:09.673Z"

	// 解析时间字符串，指定输入格式
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		fmt.Println("解析时间错误:", err)
		return "", err
	}

	// 格式化为目标格式
	newTimeStr := t.Format("2006-01-02 15:04:05")
	return newTimeStr, nil
}

func createMarkDown(title, color string, alert Alert) string {
	var markdown bytes.Buffer

	// Construct the Markdown string
	markdown.WriteString(fmt.Sprintf("# <font color=%s>%s</font>\n\n", color, title))
	//markdown.WriteString("## Items\n\n")
	markdown.WriteString(fmt.Sprintf("- summary: %s \n", alert.Annotations["summary"]))
	for key, item := range alert.Labels {
		markdown.WriteString(fmt.Sprintf("- %s: %s\n", key, item))
	}
	startsAt, err := timeFormat(alert.StartsAt)
	if err != nil {
		logger.Error("create Markdown err")
	}
	markdown.WriteString(fmt.Sprintf("- StartsAt: %s \n", startsAt))

	return markdown.String()
}

func createText(title string, alert Alert) string {
	var textContent bytes.Buffer
	textContent.WriteString(fmt.Sprintf("content: %s\n", title))
	textContent.WriteString(fmt.Sprintf("summary: %s \n", alert.Annotations["summary"]))
	for key, item := range alert.Labels {
		textContent.WriteString(fmt.Sprintf("%s: %s\n", key, item))
	}
	startsAt, err := timeFormat(alert.StartsAt)
	if err != nil {
		logger.Error("create text err")
	}
	textContent.WriteString(fmt.Sprintf("StartsAt: %s \n", startsAt))

	return textContent.String()
}

```

### apps/dingding.go

将发生报警代码提取到单独的文件

```go
package apps

import (
	"alert_gateway/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func sendMsg(message, messageType string) (string, error) {
	//secret := "SEC914ce8b70b3f05caa6b221d8da4d58b886bcd8baea8b51bd0a0e163460313a9b"
	token := "37224eaeafda63f1d98a7daca8b2b3f591d24e713f6109053678d94109482f01"
	url := "https://oapi.dingtalk.com/robot/send?access_token=" + token

	var sendDataBytes []byte
	var err error
	if messageType == "text" {
		sendData := TextMessage{
			MsgType: "text",
			Text: Text{
				Content: message,
			},
		}
		sendDataBytes, err = json.Marshal(sendData)
	} else {
		sendData := MarkdownMessage{
			MsgType: "markdown",
			Markdown: Markdown{
				Title: "消息",
				Text:  message,
				Theme: "white",
			},
		}
		sendDataBytes, err = json.Marshal(sendData)
	}

	if err != nil {
		logger.Errorf("Failed to marshal JSON: %v", err)
		return "", err
	}
	logger.Debug(string(sendDataBytes))

	reqBody := bytes.NewBuffer(sendDataBytes)

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
		return "", err
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Failed to send request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// 获取发送后的相应
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Failed to read response body: %v", err)
		return "", err
	}
	// 处理响应
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Message sent successfully")
		//fmt.Println(string(respBody))
		logger.Debugf("发送钉钉相应 %s", string(respBody))
		return string(respBody), nil
	} else {
		fmt.Printf("Failed to send message, status code: %d\n", resp.StatusCode)
		//fmt.Println(string(respBody))
		logger.Debugf("发送钉钉相应 %s", string(respBody))
		return string(respBody), nil
	}
}

```

### apps/web.go

将get方法中的帮助信息放到静态msg.txt文件中，放在项目根目录下

```go
package apps

import (
	"alert_gateway/config"
	"alert_gateway/logger"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)


// 启动应用
func Run(cfg *config.Config) {
	logger.Info("info log")
	logger.Error("error log")
	logger.Debug("debug log")

	appConfig := cfg.App
	addr := fmt.Sprintf("%s:%v", appConfig["listen"], appConfig["port"])
	logger.Infof("Server running on %s", addr)

	server := http.Server{
		Addr: addr,
	}
	http.HandleFunc("/", app.index)
	if err := server.ListenAndServe(); err != nil {
		logger.Errorf("Server error: %v", err)
	}
}

// type IndexData struct {
//  Title string `json:"tile"`
//  Desc  string `json:"desc"`
// }

// 报警结构体中的Alerts数组内的报警内容
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// 定义alertmanager发送的报警结构体
type AlertData struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
}

// 钉钉Markdown消息结构
type MarkdownMessage struct {
	MsgType  string   `json:"msgtype"`
	Markdown Markdown `json:"markdown"`
}

// 钉钉文本消息结构
type TextMessage struct {
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
}

// Markdown内容结构
type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Theme string `json:"theme"`
}

// 文本内容结构
type Text struct {
	Content string `json:"content"`
}



func index(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debugf("Received request at %s", r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			handleGet(w, r)
		case http.MethodPost:
			handlePost(w, r, cfg)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// 处理get请求
func handleGet(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Handling GET request")
	helpMessage, err := ReadFileContent("msg.txt")
	if err != nil {
		logger.Error(err)
		return
	}
	// logger.Debugf("测试数据输出 %s\n", *helpMessage)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	// 用于格式化并写入数据到指定的 io.Writer 接口中。w 是 HTTP 响应写入器，实现了 io.Writer 接口
	// 将 *fileContent（文件内容的字符串值）写入到 w 中。
	// w 是 http.ResponseWriter，它用于构建 HTTP 响应。
	// fmt.Fprintln 会在写入的数据末尾添加一个换行符。
	// fmt.Fprintln(w, *helpMessage)
	// 将字符串内容转换为字节切片并写入 HTTP 响应
	w.Write([]byte(*helpMessage))
}

// 处理post请求
func handlePost(w http.ResponseWriter, r *http.Request， cfg *config.Config) {
	logger.Debug("Handling POST request")
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		logger.Errorf("Failed to read request body: %v", err)
		return
	}
	// 调试打印请求
	// logger.Debugf("----- %s ------", getTime())
	// logger.Debugf("Received body: %s \n", string(body))
	// logger.Debugf("----- %s ------", getTime())

	var alertData AlertData
	err = json.Unmarshal(body, &alertData)
	if err != nil {
		http.Error(w, "解析json数据失败", http.StatusBadRequest)
		logger.Errorf("Failed to parse JSON data: %v", err)
	}

	// 定义一个map类型的切片，存放发送钉钉消息的结果数据
	var responses []map[string]interface{}

	// 读取配置文件设置发现钉钉的消息类型
    if cfgMessageType, ok := cfg.App["messageType"].(string); ok {
		messageType = cfgMessageType
	}

	// 循环发送报警信息到钉钉
	for _, alert := range alertData.Alerts {
		var message string
		if alert.Status == "resolved" {
			if messageType == "text" {
				message = createText("恢复", alert)
			} else {
				message = createMarkDown("恢复", "#00FF00", alert)
			}
		} else {
			if messageType == "text" {
				message = createText("故障", alert)
			} else {
				message = createMarkDown("故障", "#FF0000", alert)
			}
		}

		respMsg, err := sendMsg(message, messageType)
		response := map[string]interface{}{
			"alert":   alert.Labels["instance"],
			"respMsg": respMsg,
			"error":   err,
		}

		responses = append(responses, response)

		if err != nil {
			logger.Errorf("Failed to send message: %v", err)
		} else {
			logger.Debugf("Response message: %s", respMsg)
		}
	}

	responseBody, err := json.Marshal(responses)
	if err != nil {
		http.Error(w, "Failed to marshal response JSON", http.StatusInternalServerError)
		logger.Errorf("Failed to marshal response JSON: %v", err)
		return
	}

	// Echo the response back to the client
	w.Write(responseBody)
}

```

### 修改后的目录结构

```
.
├── app.log
├── apps
│   ├── dingding.go
│   ├── utils.go
│   └── web.go
├── cmd
│   └── main.go
├── config
│   └── config.go
├── config.toml
├── go.mod
├── go.sum
├── logger
│   └── logger.go
├── test
    └── test.json
```

### 重构web.go

配置文件要通过参数逐级传入比较麻烦，配置文件可能要在多个地方被使用到，因此将配置文件做成一个上下文

apps/web.go。

- 创建一个App结构体保存配置信息
- 定义一个初始化App实例的工厂函数
- 将web.go中的和web处理相关的函数做改成App的方法，通过实例本身获得配置信息
- 在入口函数调用App工厂函数并传入配置信息

```go
package apps

import (
	"alert_gateway/config"
	"alert_gateway/logger"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// 定义应用结构体，包含配置信息，用来
type App struct {
	config *config.Config
}

// 工厂函数，接收一个配置信息作为参数，返回一个包含配置信息的App对象
// 在包外部可以调用后获得一个包含配置信息的对象后就在外部启动web应用
func NewApp(cfg *config.Config) *App {
	return &App{config: cfg}
}

// 启动应用
func (app *App) Run() {
	logger.Info("info log")
	logger.Error("error log")
	logger.Debug("debug log")

	appConfig := app.config.App
	addr := fmt.Sprintf("%s:%v", appConfig["listen"], appConfig["port"])
	logger.Infof("Server running on %s", addr)

	server := http.Server{
		Addr: addr,
	}
	http.HandleFunc("/", app.index)
	if err := server.ListenAndServe(); err != nil {
		logger.Errorf("Server error: %v", err)
	}
}

// type IndexData struct {
//  Title string `json:"tile"`
//  Desc  string `json:"desc"`
// }

// 报警结构体中的Alerts数组内的报警内容
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// 定义alertmanager发送的报警结构体
type AlertData struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
}

// 钉钉Markdown消息结构
type MarkdownMessage struct {
	MsgType  string   `json:"msgtype"`
	Markdown Markdown `json:"markdown"`
}

// 钉钉文本消息结构
type TextMessage struct {
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
}

// Markdown内容结构
type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Theme string `json:"theme"`
}

// 文本内容结构
type Text struct {
	Content string `json:"content"`
}

func (app *App) index(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("Received request at %s", r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		app.handleGet(w, r)
	case http.MethodPost:
		app.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// func index(cfg *config.Config) http.HandlerFunc {
//  return func(w http.ResponseWriter, r *http.Request) {
//      logger.Debugf("Received request at %s", r.URL.Path)

//      switch r.Method {
//      case http.MethodGet:
//          handleGet(w, r)
//      case http.MethodPost:
//          handlePost(w, r, cfg)
//      default:
//          http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//      }
//  }
// }

// 处理get请求
func (app *App) handleGet(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Handling GET request")
	helpMessage, err := ReadFileContent("msg.txt")
	if err != nil {
		logger.Error(err)
		return
	}
	// logger.Debugf("测试数据输出 %s\n", *helpMessage)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	// 用于格式化并写入数据到指定的 io.Writer 接口中。w 是 HTTP 响应写入器，实现了 io.Writer 接口
	// 将 *fileContent（文件内容的字符串值）写入到 w 中。
	// w 是 http.ResponseWriter，它用于构建 HTTP 响应。
	// fmt.Fprintln 会在写入的数据末尾添加一个换行符。
	// fmt.Fprintln(w, *helpMessage)
	// 将字符串内容转换为字节切片并写入 HTTP 响应
	w.Write([]byte(*helpMessage))
}

// 处理post请求
func (app *App) handlePost(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Handling POST request")
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		logger.Errorf("Failed to read request body: %v", err)
		return
	}
	// 调试打印请求
	// logger.Debugf("----- %s ------", getTime())
	// logger.Debugf("Received body: %s \n", string(body))
	// logger.Debugf("----- %s ------", getTime())

	var alertData AlertData
	err = json.Unmarshal(body, &alertData)
	if err != nil {
		http.Error(w, "解析json数据失败", http.StatusBadRequest)
		logger.Errorf("Failed to parse JSON data: %v", err)
	}

	// 定义一个map类型的切片，存放发送钉钉消息的结果数据
	var responses []map[string]interface{}

	// 读取配置文件设置发现钉钉的消息类型
	messageType := "markdown" // 默认消息类型
	if cfgMessageType, ok := app.config.App["messageType"].(string); ok {
		messageType = cfgMessageType
	}

	// 循环发送报警信息到钉钉
	for _, alert := range alertData.Alerts {
		var message string
		if alert.Status == "resolved" {
			if messageType == "text" {
				message = createText("恢复", alert)
			} else {
				message = createMarkDown("恢复", "#00FF00", alert)
			}
		} else {
			if messageType == "text" {
				message = createText("故障", alert)
			} else {
				message = createMarkDown("故障", "#FF0000", alert)
			}
		}

		respMsg, err := sendMsg(message, messageType, *app.config)
		response := map[string]interface{}{
			"alert":   alert.Labels["instance"],
			"respMsg": respMsg,
			"error":   err,
		}

		responses = append(responses, response)

		if err != nil {
			logger.Errorf("Failed to send message: %v", err)
		} else {
			logger.Debugf("Response message: %s", respMsg)
		}
	}

	responseBody, err := json.Marshal(responses)
	if err != nil {
		http.Error(w, "Failed to marshal response JSON", http.StatusInternalServerError)
		logger.Errorf("Failed to marshal response JSON: %v", err)
		return
	}

	// Echo the response back to the client
	w.Write(responseBody)
}

```

cmd/main.go

```go
package main

import (
	"alert_gateway/apps"
	"alert_gateway/config"
	"alert_gateway/logger"
	"flag"
	"log"
)

func main() {
	var help bool
	var configPath string

	flag.BoolVar(&help, "help", false, "show help informaction")
	flag.StringVar(&configPath, "config", "config.toml", "path to config file")

	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.InitLogger(cfg)

	switch {
	case help:
		flag.PrintDefaults()
	default:
		// 实例化应用，传入配置信息
		app := apps.NewApp(cfg)
		// 启动应用
		app.Run()
	}
}

```

