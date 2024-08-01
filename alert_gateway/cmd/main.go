package main

import (
	"alert_gateway/apps"
	"alert_gateway/config"
	"alert_gateway/logger"
	"flag"
	"log"
)

func main() {
	var help bool
	var configPath string

	flag.BoolVar(&help, "help", false, "show help informaction")
	flag.StringVar(&configPath, "config", "config.toml", "path to config file")

	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.InitLogger(cfg)

	switch {
	case help:
		flag.PrintDefaults()
	default:
		// 实例化应用，传入配置信息
		app := apps.NewApp(cfg)
		// 启动应用
		app.Run()
	}
}
