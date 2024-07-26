package main

import (
	"flag"
	"fmt"
	"prome_cpu_temperature/prometheus"
)

func main() {
	var help bool
	var port string

	flag.BoolVar(&help, "help", false, "show help imformation")
	flag.StringVar(&port, "port", "80", "port")

	flag.Parse()

	// 如果没有提供任何参，则输出使用方法
	// if len(os.Args) < 2 {
	//  fmt.Println("Usage:")
	//  flag.PrintDefaults()
	//  os.Exit(1)
	// }

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
