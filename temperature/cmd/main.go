package main

import (
	//"flag"
	"fmt"
	"os"
	"temperature"
)

func main() {
	err := temperature.GetCPUTemperature()
	if err != nil {
		fmt.Printf("Error: %T", err)
		os.Exit(1)
	}
}
