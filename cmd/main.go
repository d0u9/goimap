package main

import (
	"os"

	log "go.uber.org/zap"
)

var (
	APP_NAME    = "goimap"
	CONFIG_NAME = "goimap.yml"
	CONFIG_DIR  = []string{"$HOME/.config/", "."}
)

func main() {
	app := NewApp()
	if err := app.Run(); err != nil {
		log.S().Fatalf("run app failed: %v", err)
		os.Exit(1)
	}
}
