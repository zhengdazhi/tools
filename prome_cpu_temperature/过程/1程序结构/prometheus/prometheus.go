package prometheus

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"prome_cpu_temperature/logutil"
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
	//logutil.LogDebug("Start HTTP server on %s", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}

// 收集并设置CPU指标
func collectAndSetMetrics() error {
	// 死循环不断更新数据
	for {
		// 通过temperature包中的接口获取温度数据
		cpuData, err := temperature.GetCPUTemperature()
		if err != nil {
			fmt.Printf("Error: %T", err)
			os.Exit(1)
		}

		// 模拟数据，这里应该用实际从temperature包获取的数据
		// cpuData := temperature.CPUData{
		// 	CPUCores:       4,
		// 	MaxTemperature: 75.0,
		// 	MinTemperature: 35.0,
		// 	AvgTemperature: 55.0,
		// }

		// 设置Prometheus指标值
		cpuCoreCount.Set(float64(cpuData.CPUCores))
		cpuCoreTemperatureMax.Set(cpuData.MaxTemperature)
		cpuCoreTemperatureMin.Set(cpuData.MinTemperature)
		cpuCoreTemperatureAvg.Set(cpuData.AvgTemperature)
		logutil.LogDebug("Collected CPU metrics: cores=%d, max=%.2f, min=%.2f, avg=%.2f",
			cpuData.CPUCores,
			cpuData.MaxTemperature,
			cpuData.MinTemperature,
			cpuData.AvgTemperature,
		)

		time.Sleep(10 * time.Second)
	}

}
