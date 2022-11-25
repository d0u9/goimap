package main

import (
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

func (cmd *DaemonCmd) Main(config *DaemonConfig, appConfig *AppConfig) error {
	slog.Infof("%+v\n", config)

	return nil
}
