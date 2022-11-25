package main

import (
	"os"

	"github.com/gookit/slog"
)

const APP_NAME = "goimap"

func main() {
	app := NewApp()
	if err := app.Run(); err != nil {
		slog.Fatal(err)
		os.Exit(1)
	}
}
