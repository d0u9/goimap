package main

import (
	// "github.com/gookit/slog"
	"github.com/spf13/pflag"
)

type AppConfig struct {
	ConfigFile string
}

type App struct {
	Command
}

type Command interface {
	Execute() error
}

/*
func BindRootPersistentFlags(flags *pflag.FlagSet, appConfig *AppConfig) {
    flags.StringVarP(&appConfig.ConfigFile, "config", "c", "", "config file")
}

func BindDaemonCmdFlags(flags *pflag.FlagSet, daemonConfig *DaemonConfig) {
    flags.BoolVarP(&daemonConfig.Deamon, "deamon", "d", false, "run in background, by default goimap runs in foreground")
}
*/

func AddAppFlags(flags *pflag.FlagSet, config *AppConfig) {
	flags.StringVarP(&config.ConfigFile, "config", "c", "", "config file")
}

func NewApp() *App {
	var appConfig = &AppConfig{}

	syncCmd := NewSyncCmd(appConfig)
	daemonCmd := NewDaemonCmd(appConfig)

	// The root command is default to sync
	rootCmd := NewSyncCmd(appConfig)
	rootCmd.Use = "goimap"
	rootCmd.Short = "an imap client"
	rootCmd.AddCommand(&syncCmd.Command)
	rootCmd.AddCommand(&daemonCmd.Command)
	AddAppFlags(rootCmd.PersistentFlags(), appConfig)

	return &App{
		Command: rootCmd,
	}
}

func (app *App) Run() error {
	return app.Execute()
}
