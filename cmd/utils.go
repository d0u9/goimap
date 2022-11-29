package main

import (
	"os"
	"os/signal"

	"github.com/spf13/viper"
	log "go.uber.org/zap"
)

func ReadConfigFile(file string) error {

	if file != "" {
		viper.SetConfigFile(file)
	} else {
		viper.SetConfigName(CONFIG_NAME)
		viper.SetConfigType("yaml")
		for _, path := range CONFIG_DIR {
			viper.AddConfigPath(path)
		}
	}

	return viper.ReadInConfig()
}

func InitFetchers() {
	activeAccounts := viper.GetStringSlice("general.accounts")

	log.S().Infoln(activeAccounts)
}

func newOSWatcher(sigs ...os.Signal) chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)

	return sigChan
}
