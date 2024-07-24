package cpu_temperature_test

import (
	"runtime"
	"temperature"
	"testing"
)

func TestFetchCPUTemperature(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping test on non-Windows systems")
	}

	getter := temperature.WindowsTemperatureGetter{}
	if err := getter.FetchCPUTemperature(); err != nil {
		t.Errorf("Failed to fetch CPU temperature: %v", err)
	}
}
