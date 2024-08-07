package temperature

import (
	"fmt"
	"prome_cpu_temperature/logutil"
	"runtime"
	"time"
)

// CPUData 用于存储CPU温度数据
type CPUData struct {
	CPUCores       int
	MaxTemperature float64
	MinTemperature float64
	AvgTemperature float64
}

// 定义一个接口获取cpu温度
type CPUTemperatureGetter interface {
	FetchCPUTemperature() (*CPUData, error)
}

// 工厂函数，根据系统类型返回对应的CPUTemperatureGetter实现
func NewCPUTemperatureGetter() (CPUTemperatureGetter, error) {
	// 根据不同的系统架构返回对应的底层实现
	switch runtime.GOOS {
	case "windows":
		return WindowsTemperatureGetter{}, nil
	case "linux":
		return LinuxTemperatureGetter{}, nil
	default:
		return nil, fmt.Errorf("unsupported opeaating system %s", runtime.GOOS)
	}
}

// 封装一个函数让外部调用
func GetCPUTemperature() (*CPUData, error) {
	// 创建收集器
	cpuTemperature, err := NewCPUTemperatureGetter()
	if err != nil {
		return nil, fmt.Errorf("error creating CPU temperature getter: %w", err)
	}
	logutil.LogDebug("创建收集器成")
	// 获取数据
	data, err := cpuTemperature.FetchCPUTemperature()
	if err != nil {
		return nil, fmt.Errorf("error getting CPU temperatue: %w", err)
	}
	logutil.LogDebug("收集数据完成")
	currentTime := time.Now()
	formatted := currentTime.Format("2006-01-02 15:04:05")
	logutil.LogDebug("######## %s ######\n", formatted)
	logutil.LogDebug("CPU Cores: %d\n", data.CPUCores)
	logutil.LogDebug("Max Temperature: %.2f°C\n", data.MaxTemperature)
	logutil.LogDebug("Min Temperature: %.2f°C\n", data.MinTemperature)
	logutil.LogDebug("Avg Temperature: %.2f°C\n", data.AvgTemperature)
	return data, nil
}
