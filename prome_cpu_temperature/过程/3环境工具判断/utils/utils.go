package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"prome_cpu_temperature/logutil"
	"runtime"
	"time"
)

// 检测依赖工具是否配置齐全
func CheckTools() (bool, error) {
	switch runtime.GOOS {
	case "windows":
		logutil.LogDebug("Enviroment is Windows")
		return checkWindowsTools()
	case "linux":
		logutil.LogDebug("Enviroment is Linux")
		return checkLinuxTools()
	default:
		return false, fmt.Errorf("unsupported opeaating system %s", runtime.GOOS)
	}
}

// 检查Windows系统所需工具
func checkWindowsTools() (bool, error) {
	exe, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("error getting executable path: %w", err)
	}
	dir := filepath.Dir(exe)
	pwd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("error getting working directory: %w", err)
	}

	directories := []string{
		filepath.Join(dir, "tools", "OpenHardwareMonitor", "OpenHardwareMonitor.exe"),
		filepath.Join(pwd, "tools", "OpenHardwareMonitor", "OpenHardwareMonitor.exe"),
	}

	for _, filePath := range directories {
		if _, err := os.Stat(filePath); err == nil {
			logutil.LogDebug("Found OpenHardwareMonitor.exe at: ", filePath)
			// 开启OpenHardwareMonitor程序
			go runOpenHardwareMonitor(filePath)
			maxRetries := 10
			timeout := 1 * time.Second
			// 检测OpenHardwareMonitor是否正常开启
			err := checkOpenHardwareMonitor("http://127.0.0.1:8085/data.json", maxRetries, timeout)
			if err != nil {
				fmt.Printf("测试请求超时 %s \n", err)
				return false, err
			}
			return true, nil
		}
	}
	return false, errors.New("failed to find OpenHardwareMonitor.exe")
}

func runOpenHardwareMonitor(filePath string) {
	fmt.Printf("OpenHardwareMonitor dir: %s \n", filePath)
	cmd := exec.Command(filePath)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// 判断OpenHardwareMonitor是否正常开启
func checkOpenHardwareMonitor(url string, maxRetries int, timeout time.Duration) error {
	fmt.Println("等待OpenHardwareMonitor开启")
	client := &http.Client{
		Timeout: timeout,
	}

	for i := 0; i < maxRetries; i++ {
		// 创建带有超时的上下文
		time.Sleep(timeout)
		fmt.Printf("请求 %d 次\n", i)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			if os.IsTimeout(err) {
				fmt.Printf("Attempt %d: Request timed out after %v\n", i+1, timeout)
				continue
			} else {
				fmt.Errorf("attempt %d: failed to make request: %v", i+1, err)
				continue
			}
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Printf("Attempt %d: Request succeeded\n", i+1)
			return nil
		} else {
			fmt.Printf("Attempt %d: Request failed with status: %s\n", i+1, resp.Status)
		}
	}

	return errors.New("all attempts failed due to timeout")
}

// 检查Linux系统所需工具
func checkLinuxTools() (bool, error) {
	if info, err := os.Lstat("/usr/bin/sensors"); err == nil && !info.IsDir() {
		return true, nil
	}
	return false, errors.New("sensors tool not found")
}
