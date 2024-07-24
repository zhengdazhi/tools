package temperature

import (
	"fmt"
	"runtime"
)

// 定义一个接口获取cpu温度
type CPUTemperatureGetter interface {
	FetchCPUTemperature() error
}

// 工厂函数，根据系统类型返回对应的CPUTemperatureGetter实现
func NewCPUTemperatureGetter() (CPUTemperatureGetter, error) {
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
func GetCPUTemperature() error {
	cpuTemperature, err := NewCPUTemperatureGetter()
	if err != nil {
		return fmt.Errorf("error creating CPU temperature getter: %w", err)
	}
	if err := cpuTemperature.FetchCPUTemperature(); err != nil {
		return fmt.Errorf("error getting CPU temperatue: %w", err)
	}
	return nil
}
