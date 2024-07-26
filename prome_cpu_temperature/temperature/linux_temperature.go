package temperature

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// linux 系统实现获取cpu温度接口
type LinuxTemperatureGetter struct {
}

func (l LinuxTemperatureGetter) FetchCPUTemperature() (*CPUData, error) {
	fmt.Println("linux cpu temperature")
	cpu_tem, err := getCpuTem()
	if err != nil {
		return nil, err
	}
	avg_tem, err := getPackageTem(cpu_tem)
	if err != nil {
		return nil, err
	}
	fmt.Printf("linux cpu温度: %0.2f°C\n", avg_tem)
	// 以下是模拟数据，用来返回
	return &CPUData{
		CPUCores:       4,
		MaxTemperature: 75.0,
		MinTemperature: 35.0,
		AvgTemperature: 55.0,
	}, nil
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

// 收集cpu封装温度信息
func getPackageTem(cpu_data []string) (float64, error) {
	// 定义正则匹配规则
	var temperatureRegex = regexp.MustCompile(`\+\d+\.\d+°C`)
	var tems []string
	for _, line := range cpu_data {
		if strings.HasPrefix(line, "Package id") {
			//fmt.Println(line)
			match := temperatureRegex.FindString(line)
			if match != "" {
				temperatureStr := strings.TrimSuffix(strings.TrimPrefix(match, "+"), "°C") // 去除前缀和后缀，得到温度字符串
				//fmt.Println(temperatureStr)
				tems = append(tems, temperatureStr)
			}
		}
	}
	if len(tems) == 0 {
		return 0, fmt.Errorf("没有cpu封装温度数据")
	}
	// 计算多个物理cpu的平均温度
	if len(tems) > 1 {
		var sum_tem float64
		for _, tem := range tems {
			value, err := strconv.ParseFloat(tem, 64)
			if err != nil {
				return 0, fmt.Errorf("转换温度数据错误: ", err)
			}
			sum_tem += value
		}
		avg_tem := sum_tem / float64(len(tems))
		return avg_tem, nil
	} else {
		// 单核cpu直接返回结果
		value, err := strconv.ParseFloat(tems[0], 64)
		if err != nil {
			return 0, fmt.Errorf("转换温度数据错误: ", err)
		}
		return value, nil
	}
}
