package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	//"github.com/go-wmi/wmi"
	//"github.com/StackExchange/wmi"
)

func main() {
	// 原始命令 wmic /namespace:\\root\wmi PATH MSAcpi_ThermalZoneTemperature get CurrentTemperature
	cmd := exec.Command("wmic", "/namespace:\\\\root\\wmi", "PATH", "MSAcpi_ThermalZoneTemperature", "get", "CurrentTemperature")
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	// 启动命令
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	// 读取命令的输出
	data, err := io.ReadAll(out)
	if err != nil {
		log.Fatal(err)
	}

	// 等待命令完成
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	// 将输出转换为字符串并去除多余的空白
	output := strings.TrimSpace(string(data))

	// 打印原始输出
	// fmt.Println("原始输出：\n", output)

	// 解析输出（假设输出包含标题和数据）
	lines := strings.Split(output, "\n")
	if len(lines) > 1 {
		// 从第二行获取温度数据
		temperatureRaw := strings.TrimSpace(lines[1])
		temperature := parseTemperature(temperatureRaw)
		fmt.Printf("当前温度：%.2f°C\n", temperature)
	} else {
		fmt.Println("未找到温度数据")
	}

}

// parseTemperature 将原始温度字符串转换为摄氏度的 float64 值
func parseTemperature(raw string) float64 {
	// 温度值以千分之一开尔文（K）为单位报告，转换为摄氏度
	// 转换公式是 (Temperature(K) - 273.15)
	// 例如：如果原始值是 3000，则表示 300.0K -> 27.85°C
	if temp, err := strconv.ParseFloat(raw, 64); err == nil {
		return (temp/10 - 273.15)
	}
	return 0
}
