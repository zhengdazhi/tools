package main

import (
	"flag"
	"fmt"
	"prome_cpu_temperature/logutil"
	"prome_cpu_temperature/prometheus"
)

func main() {
	var help bool
	var port string
	var debug bool

	flag.BoolVar(&help, "help", false, "show help imformation")
	flag.StringVar(&port, "port", "80", "port")
	flag.BoolVar(&debug, "debug", false, "enable debug mode")

	flag.Parse()

	// 启用或者禁用debug日志
	logutil.SetDebug(debug)

	switch {
	case help:
		flag.PrintDefaults()
	case port != "":
		fmt.Printf("port is %s \n", port)
		prometheus.Run(port)
	default:
		flag.PrintDefaults()
	}
}
