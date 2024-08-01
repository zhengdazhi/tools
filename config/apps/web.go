package apps

import (
	"config/config"
	"config/logger"
	"encoding/json"
	"fmt"
	"net/http"
)

func Run(cfg *config.Config) {
	logger.Info("web")
	appConfig := cfg.App
	addr := fmt.Sprintf("%s:%v", appConfig["listen"], appConfig["port"])
	//addr := appConfig["listen"] + ":" + appConfig["port"]
	server := http.Server{
		Addr: addr,
	}
	http.HandleFunc("/", index)
	if err := server.ListenAndServe(); err != nil {
		//log.Println(err)
		logger.Error(err)
	}
}

type IndexData struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

func index(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("Received request at: %s", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	var indexData IndexData
	indexData.Title = "go博客"
	indexData.Desc = "现在是入门教程"
	jsonStr, _ := json.Marshal(indexData)
	w.Write(jsonStr)
}
