package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type SyncConfig struct {
	Sock       string
	Standalone bool
}

func (conf *SyncConfig) addFlags(flags *pflag.FlagSet) {
	flags.StringVar(&conf.Sock, "sock", "", "daemon sock file")
	flags.BoolVar(&conf.Standalone, "standalone", false, "fetch directly, do not use daemon")
}

type SyncCmd struct {
	cobra.Command
	Config SyncConfig
}

func NewSyncCmd(appConfig *AppConfig) *SyncCmd {
	var config SyncConfig
	var syncCmd = SyncCmd{}
	var cmd = cobra.Command{
		Use:   "sync",
		Short: "sync local folder with remote imap server",
		Run: func(cmd *cobra.Command, args []string) {
			syncCmd.Main(&config, appConfig)
		},
	}

	config.addFlags(cmd.Flags())
	syncCmd.Command = cmd

	return &syncCmd
}

func (cmd *SyncCmd) Main(config *SyncConfig, appConfig *AppConfig) error {

	return nil
}
