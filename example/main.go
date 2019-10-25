package main

import (
	zap "github.com/928799934/zap"
	"time"
)

func main() {
	logger := zap.LoadConfiguration("./logconfig.json")
	defer zap.Close()
	log := logger.Sugar()

	for {
		log.Error("error")
		log.Info("info")
		time.Sleep(5 * time.Second)
	}
}
