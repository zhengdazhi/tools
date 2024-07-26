package prometheus

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"prome_cpu_temperature/temperature"

	"github.com/prometheus/client_golang/prometheus"          // Prometheus客户端库
	"github.com/prometheus/client_golang/prometheus/promhttp" // Prometheus HTTP处理器
)

// 定义全局变量用于存储Prometheus指标
var (
	//prometheus.NewGauge 是 Prometheus 客户端库中的一个函数，用于创建一个新的 Gauge 指标。
	//Gauge 是一种指标类型，可以表示一个瞬时值，它可以增加、减少或设置到任意值，适合用于表示当前状态，比如温度、内存使用、当前连接数等。
	cpuCoreCount = prometheus.NewGauge(prometheus.GaugeOpts{
		// prometheus.GaugeOpts 是一个结构体，用于配置 Gauge 指标的各种属性。
		// 该结构体包含多个字段，如 Name（指标名称）、Help（指标帮助信息）等。
		Name: "cpu_core_count",
		Help: "Number of CPU cores",
	})
	cpuCoreTemperatureMax = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_core_temperature_max",
		Help: "Maximum temperature of all CPU cores",
	})
	cpuCoreTemperatureMin = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_core_temperature_min",
		Help: "Minimum temperature of all CPU cores",
	})
	cpuCoreTemperatureAvg = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_core_temperature_avg",
		Help: "Average temperature of all CPU cores",
	})
)

// 定义结构体用于存储CPU指标数据
type cpuMetrics struct {
	cpuCoreCount          int     // CPU核心数量
	cpuCoreTemperatureMax float64 // 所有CPU核心的最高温度
	cpuCoreTemperatureMin float64 // 所有CPU核心的最低温度
	cpuCoreTemperatureAvg float64 // 所有CPU核心的平均温度
}

// 初始化注册Prometheus指标
func init() {
	// 使用 prometheus.MustRegister 函数将这个指标注册到 Prometheus 的默认注册表中
	prometheus.MustRegister(cpuCoreCount)
	prometheus.MustRegister(cpuCoreTemperatureMax)
	prometheus.MustRegister(cpuCoreTemperatureMin)
	prometheus.MustRegister(cpuCoreTemperatureAvg)
}

// 启动主程序
func Run(port string) {
	var wg sync.WaitGroup
	wg.Add(2)
	// 启动http服务
	go func() {
		defer wg.Done()
		startHttp(port)
	}()
	// 收集更新数据
	go func() {
		defer wg.Done()
		err := collectAndSetMetrics()
		if err != nil {
			fmt.Printf("Error collecting metrics: %v\n", err)
			os.Exit(1)
		}
	}()
	wg.Wait()
}

// 启动prometheus http服务
func startHttp(port string) {
	httpAddr := ":" + port
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting HTTP server on %s", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}

// 收集并设置CPU指标
func collectAndSetMetrics() error {
	// 死循环不断更新数据
	for {
		cpuData, err := temperature.GetCPUTemperature()
		if err != nil {
			fmt.Printf("Error: %T", err)
			os.Exit(1)
		}

		// 模拟数据，这里应该用实际从temperature包获取的数据
		// cpuData := temperature.CPUData{
		//  CPUCores:       4,
		//  MaxTemperature: 75.0,
		//  MinTemperature: 35.0,
		//  AvgTemperature: 55.0,
		// }

		// 设置Prometheus指标值
		cpuCoreCount.Set(float64(cpuData.CPUCores))
		cpuCoreTemperatureMax.Set(cpuData.MaxTemperature)
		cpuCoreTemperatureMin.Set(cpuData.MinTemperature)
		cpuCoreTemperatureAvg.Set(cpuData.AvgTemperature)

		// 每10秒收集一次数据
		time.Sleep(10 * time.Second)
	}

}

// 检测依赖工具是否配置齐全
func checkTools() (bool, error) {
	switch runtime.GOOS {
	case "windows":
		fmt.Println("env is windows")
		systemArch := "windows"
		ok, err := getExeDir(systemArch)
		if ok {
			return true, nil
		} else {
			return false, err
		}
	case "linux":
		fmt.Println("evn is linux")
		systemArch := "linux"
		ok, err := getExeDir(systemArch)
		if ok {
			return true, nil
		} else {
			return false, err
		}
	default:
		return false, fmt.Errorf("unsupported opeaating system %s", runtime.GOOS)
	}
}

// 获取
func getExeDir(systemArch string) (bool, error) {
	if systemArch == "windows" {
		exe, err := os.Executable()
		if err != nil {
			panic(err)
		}
		dir := filepath.Dir(exe)
		fmt.Println("Executable directory: ", dir)
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting working directory:", err)
		}
		fmt.Println("Current working directory:", pwd)
		filePath := filepath.Join(dir, "/tools/OpenHardwareMonitor/OpenHardwareMonitor.exe")
		if _, err := os.Stat(filePath); err == nil {
			// 没有错误发生，文件存在
			if info, err := os.Lstat(filePath); err == nil && !info.IsDir() {
				// 确保它是一个文件而不是目录
				return true, nil
			} else {
				return false, err
			}
		} else {
			return false, err
		}

		filePath2 := filepath.Join(pwd, "/tools/OpenHardwareMonitor/OpenHardwareMonitor.exe")
		if _, err := os.Stat(filePath2); err == nil {
			// 没有错误发生，文件存在
			if info, err := os.Lstat(filePath2); err == nil && !info.IsDir() {
				// 确保它是一个文件而不是目录
				return true, nil
			} else {
				return false, err
			}
		} else {
			return false, err
		}
	}
	if systemArch == "linux" {
		if info, err := os.Lstat("/usr/bin/sensors"); err == nil && !info.IsDir() {
			return true, nil
		}
	}

	return false, errors.New("没有找到OpenHardwareMonitor.exe程序")
}
