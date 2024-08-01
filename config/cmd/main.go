package main

import (
	"config/apps"
	"config/config"
	"config/logger"
	"flag"
	"log"
)

func main() {
	var help bool
	var configPath string
	flag.BoolVar(&help, "help", false, "show help informaction")
	flag.StringVar(&configPath, "config", "config.toml", "path to the config file")

	flag.Parse()

	// 加载配置文件
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志模块
	logger.InitLogger(cfg)

	switch {
	case help:
		flag.PrintDefaults()
	default:
		apps.Run(cfg)
	}

	// 使用加载的配置和其他命令行标志
	// fmt.Println("Title:", cfg.Title)
	// fmt.Println("Log Path:", cfg.Log.Path)
	// fmt.Println("Log Level:", cfg.Log.Level)
	// fmt.Println("App Port:", cfg.App.Port)
	// fmt.Println("Database Server:", cfg.Database.Server)
	// fmt.Println("Database Port:", cfg.Database.Port)
	// fmt.Println("Database User:", cfg.Database.User)
	// fmt.Println("Database Password:", cfg.Database.Password)
}
