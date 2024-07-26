package cpu_temperature_test

import (
	"cpu_temperature"
	"testing"
)

func TestFetchCPUTemperature(t *testing.T) {
	getter := cpu_temperature.LinuxTemperatureGetter{}
	if err := getter.FetchCPUTemperature(); err != nil {
		t.Errorf("Failed to fetch CPU temperature: %v", err)
	}
}
