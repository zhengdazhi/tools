package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// 定义用于解析 JSON 的结构体
type Node struct {
	ID       int    `json:"id"`
	Text     string `json:"Text"`
	Min      string `json:"Min"`
	Value    string `json:"Value"`
	Max      string `json:"Max"`
	ImageURL string `json:"ImageURL"`
	Children []Node `json:"Children"`
}

// 定义 CPUData 结构体
type CPUData struct {
	CPUCores       int
	MaxTemperature float64
	MinTemperature float64
	AvgTemperature float64
}

// 递归遍历节点以提取 CPU 温度信息
func extractCPUTemperatures(node Node, temperatures *[]float64) {
	if node.Text == "Temperatures" {
		for _, child := range node.Children {
			if strings.HasPrefix(child.Text, "CPU Core #") || child.Text == "CPU Package" {
				if temp, err := strconv.ParseFloat(strings.TrimSuffix(child.Value, " °C"), 64); err == nil {
					*temperatures = append(*temperatures, temp)
				}
			}
		}
	}

	for _, child := range node.Children {
		extractCPUTemperatures(child, temperatures)
	}
}

// 计算 CPU 温度的统计数据
func calculateCPUData(temperatures []float64) CPUData {
	var data CPUData
	if len(temperatures) == 0 {
		return data
	}

	data.CPUCores = len(temperatures)
	sum := 0.0
	data.MinTemperature = temperatures[0]
	data.MaxTemperature = temperatures[0]

	for _, temp := range temperatures {
		if temp > data.MaxTemperature {
			data.MaxTemperature = temp
		}
		if temp < data.MinTemperature {
			data.MinTemperature = temp
		}
		sum += temp
	}

	data.AvgTemperature = sum / float64(data.CPUCores)
	return data
}

func main() {
	// 假设 JSON 数据来自于本地文件或者 HTTP 请求
	url := "http://localhost:8085/data.json" // 修改为实际的 JSON 数据来源
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching data: %v\n", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v\n", err)
	}

	var root Node
	if err := json.Unmarshal(body, &root); err != nil {
		log.Fatalf("Error parsing JSON: %v\n", err)
	}

	// 提取 CPU 温度信息
	var temperatures []float64
	extractCPUTemperatures(root, &temperatures)

	// 计算 CPU 温度统计数据
	cpuData := calculateCPUData(temperatures)

	// 打印结果
	fmt.Printf("CPU Cores: %d\n", cpuData.CPUCores)
	fmt.Printf("Max Temperature: %.2f°C\n", cpuData.MaxTemperature)
	fmt.Printf("Min Temperature: %.2f°C\n", cpuData.MinTemperature)
	fmt.Printf("Avg Temperature: %.2f°C\n", cpuData.AvgTemperature)
}
