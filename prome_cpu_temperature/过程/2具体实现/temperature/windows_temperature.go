package temperature

import (
	"encoding/json"
	"io"

	//"io/ioutil"
	"log"
	"net/http"
	"prome_cpu_temperature/logutil"
	"strconv"
	"strings"
)

// windows系统实现获取cpu温度接口
type WindowsTemperatureGetter struct {
}

func (w WindowsTemperatureGetter) FetchCPUTemperature() (*CPUData, error) {
	logutil.LogDebug("使用windows系统")
	cpuData, err := openHardwareMonitorGetter()
	if err != nil {
		log.Println("获取温度数据失败\n", err)
		return nil, err
	}
	// 打印结果
	// fmt.Println("--------------------------------------------------")
	// fmt.Printf("CPU Cores: %d\n", cpuData.CPUCores)
	// fmt.Printf("Max Temperature: %.2f°C\n", cpuData.MaxTemperature)
	// fmt.Printf("Min Temperature: %.2f°C\n", cpuData.MinTemperature)
	// fmt.Printf("Avg Temperature: %.2f°C\n", cpuData.AvgTemperature)
	// fmt.Println("--------------------------------------------------")
	return &cpuData, err

	// 以下是模拟数据，用来返回测试
	// return &CPUData{
	//  CPUCores:       4,
	//  MaxTemperature: 75.0,
	//  MinTemperature: 35.0,
	//  AvgTemperature: 55.0,
	// }, nil

}

func openHardwareMonitorGetter() (CPUData, error) {
	logutil.LogDebug("使用openHardwareMonitor获取数据")
	// 假设 JSON 数据来自于本地文件或者 HTTP 请求
	url := "http://localhost:8085/data.json" // 修改为实际的 JSON 数据来源
	resp, err := http.Get(url)
	if err != nil {
		//log.Fatalf("Error fetching data: %v\n", err)
		return CPUData{}, err
	}
	defer resp.Body.Close()

	//body, err := ioutil.ReadAll(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v\n", err)
		return CPUData{}, err
	}

	var root Node
	if err := json.Unmarshal(body, &root); err != nil {
		//log.Fatalf("Error parsing JSON: %v\n", err)
		return CPUData{}, err
	}

	// 提取 CPU 温度信息
	var temperatures []float64
	extractCPUTemperatures(root, &temperatures)

	// 计算 CPU 温度统计数据
	cpuData := calculateWindowsCPUData(temperatures)
	// 打印结果
	// logutil.LogDebug("--------------------------------------------------")
	// logutil.LogDebug("CPU Cores: %d\n", cpuData.CPUCores)
	// logutil.LogDebug("Max Temperature: %.2f°C\n", cpuData.MaxTemperature)
	// logutil.LogDebug("Min Temperature: %.2f°C\n", cpuData.MinTemperature)
	// logutil.LogDebug("Avg Temperature: %.2f°C\n", cpuData.AvgTemperature)
	// logutil.LogDebug("--------------------------------------------------")
	return cpuData, nil
}

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

// 递归遍历节点以提取 CPU 温度信息
func extractCPUTemperatures(node Node, temperatures *[]float64) {
	if node.Text == "Temperatures" {
		for _, child := range node.Children {
			//if strings.HasPrefix(child.Text, "CPU Core #") || child.Text == "CPU Package" {
			// 只收集cpu核心温度，不收集封装温度
			if strings.HasPrefix(child.Text, "CPU Core #") {
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
func calculateWindowsCPUData(temperatures []float64) CPUData {
	logutil.LogDebug("统计cpu多核平均温度")
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
