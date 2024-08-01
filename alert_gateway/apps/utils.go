package apps

import (
	"alert_gateway/logger"
	"bytes"
	"fmt"
	"io/ioutil"
	"time"
)

// ReadFileContent 函数读取文件内容，并将其保存到字符串指针中
func ReadFileContent(filePath string) (*string, error) {
	// 读取文件内容
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 将内容转换为字符串，并保存到变量中
	fileContent := string(content)

	// 返回字符串指针
	return &fileContent, nil
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
