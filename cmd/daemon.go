package main

import (
	"os"
	"syscall"

	"goimap/pkg/account"

	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	log "go.uber.org/zap"
)

type DaemonConfig struct {
	Foreground bool
}

func (conf *DaemonConfig) addFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&conf.Foreground, "foreground", "f", false, "fetch directly, do not use daemon")
}

type DaemonCmd struct {
	cobra.Command
	Config DaemonConfig
}

func NewDaemonCmd(appConfig *AppConfig) *DaemonCmd {
	var config DaemonConfig
	var daemonCmd = DaemonCmd{}

	var cmd = cobra.Command{
		Use:   "daemon",
		Short: "run as a daemon",
		Run: func(cmd *cobra.Command, args []string) {
			daemonCmd.Main(&config, appConfig)
		},
	}

	config.addFlags(cmd.Flags())
	daemonCmd.Command = cmd

	return &daemonCmd
}

func (cmd *DaemonCmd) Main(daemonConfig *DaemonConfig, appConfig *AppConfig) {

	if err := ReadConfigFile(appConfig.ConfigFile); err != nil {
		slog.Fatalf("%v\n", err)
	}

	manager := account.NewAccountManager()
	if err := manager.InitFromViper(); err != nil {
		slog.Fatalf("%v\n", err)
	}

	manager.SyncAllAccounts()

	done := make(chan any, 1)
	sigs := newOSWatcher(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	hits := 2
	for {
		select {
		case s := <-sigs:
			switch s {
			case syscall.SIGINT:
				log.S().Infof("Received signal \"%v\", shutting down.", s)
				if hits == 2 {
					go func() {
						manager.Shutdown()
						manager.WaitAll()
						done <- 1
					}()
				} else if hits == 0 {
					os.Exit(1)
				}
				hits -= 1
			}
		case <-done:
			os.Exit(0)
		}
	}
}
