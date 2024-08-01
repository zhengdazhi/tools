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
