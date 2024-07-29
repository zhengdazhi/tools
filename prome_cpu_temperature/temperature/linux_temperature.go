package temperature

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// linux 系统实现获取cpu温度接口
type LinuxTemperatureGetter struct {
}

// 实现temperature中的接口
func (l LinuxTemperatureGetter) FetchCPUTemperature() (*CPUData, error) {
	fmt.Println("linux cpu temperature")
	// 采集原始数据
	cpu_tem, err := getCpuTem()
	// for _, line := range cpu_tem {
	// 	fmt.Println(line)
	// }
	if err != nil {
		return nil, err
	}
	// 统计温度数据
	cpuData, err := calculateLinuxCPUData(cpu_tem)
	if err != nil {
		return nil, err
	}
	return &cpuData, nil

	// 以下是模拟数据，用来返回
	// return &CPUData{
	// 	CPUCores:       4,
	// 	MaxTemperature: 75.0,
	// 	MinTemperature: 35.0,
	// 	AvgTemperature: 55.0,
	// }, nil
}

// 通过sensors命令获取cpu温度信息
func getCpuTem() ([]string, error) {
	cmd := exec.Command("sensors")

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(out)
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	//fmt.Println("temperature:\n", string(data))
	//cpu_data := string(data)
	cpu_data := strings.Split(string(data), "\n")
	return cpu_data, nil
}

// 收集cpu温度信息
func calculateLinuxCPUData(cpu_data []string) (CPUData, error) {
	var data CPUData
	// 定义正则匹配规则
	var temperatureRegex = regexp.MustCompile(`\+\d+\.\d+°C`)
	var tems []float64
	for _, line := range cpu_data {
		// 收集cpu 封装温度
		//if strings.HasPrefix(line, "Package id") {
		// 获取核心温度数据
		if strings.HasPrefix(line, "Core ") {
			//fmt.Println(line)
			match := temperatureRegex.FindString(line)
			if match != "" {
				temperatureStr := strings.TrimSuffix(strings.TrimPrefix(match, "+"), "°C") // 去除前缀和后缀，得到温度字符串
				//fmt.Println(temperatureStr)
				temperature, err := strconv.ParseFloat(temperatureStr, 64)
				if err != nil {
					log.Fatal("转换温度数据错误: ", err)
				}
				tems = append(tems, temperature)
			}
		}
	}
	// for _, tem := range tems {
	// 	fmt.Println(tem)
	// }
	if len(tems) == 0 {
		return data, errors.New("没有获得cpu核心温度数据")
	}
	// 获得cpu核心数
	data.CPUCores = len(tems)
	// 计算多个物理cpu的平均温度,和获取最大和最新温度
	sum := 0.0
	data.MinTemperature = tems[0]
	data.MaxTemperature = tems[0]
	for _, tem := range tems {
		if tem > data.MaxTemperature {
			data.MaxTemperature = tem
		}
		if tem < data.MinTemperature {
			data.MinTemperature = tem
		}
		sum += tem
	}
	data.AvgTemperature = sum / float64(data.CPUCores)
	return data, nil
}
