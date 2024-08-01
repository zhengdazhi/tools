package apps

import (
	"alert_gateway/config"
	"alert_gateway/logger"

	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func sendMsg(message, messageType string, cfg config.Config) (string, error) {
	baseURL := cfg.App["webhook_url"].(string)
	token := cfg.App["token"].(string)
	timestamp := getTimestamp()
	var dingdingUrl string
	if secret, ok := cfg.App["secret"].(string); ok {
		// sign := makeSign(string(timestamp), secret)
		sign := makeSign(fmt.Sprintf("%d", timestamp), secret)
		// url = url + "access_token=" + token + "&timestamp=" + string(timestamp) + "&sign=" + sign
		dingdingUrl = baseURL + "access_token=" + token + "&timestamp=" + fmt.Sprintf("%d", timestamp) + "&sign=" + sign
	} else {
		dingdingUrl = baseURL + "access_token=" + token
	}

	logger.Debugf("ding ding webhook url: %s", dingdingUrl)

	var sendDataBytes []byte
	var err error
	// 判断是文本消息还是markdown消息
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
	req, err := http.NewRequest("POST", dingdingUrl, reqBody)
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

	// 获取发送钉钉消息后的相应
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

// 使用secret对钉钉消息进行签名，返回一个签名后的字符串
func makeSign(timestamp, secret string) string {
	// 将 timestamp 和 secret 拼接成字符串
	stringToSign := strings.Join([]string{timestamp, secret}, "\n")

	// 计算 HMAC-SHA256
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	hmacCode := h.Sum(nil)

	// Base64 编码并 URL 编码
	sign := url.QueryEscape(base64.StdEncoding.EncodeToString(hmacCode))

	return sign
}

func getTimestamp() int64 {
	// 获取当前时间的时间戳（以纳秒为单位）
	now := time.Now().UnixNano()
	// 转换为毫秒级时间戳
	timestamp := now / 1000000
	//fmt.Println(timestamp)
	return timestamp
}
