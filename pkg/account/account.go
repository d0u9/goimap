package account

import (
	"context"
	"fmt"
	"time"

	"goimap/pkg/errors"

	"github.com/cenkalti/backoff/v4"
	"github.com/emersion/go-imap/client"
	"github.com/spf13/viper"
	log "go.uber.org/zap"
)

type AccountManager struct {
	Accounts map[string]*Account
}

func NewAccountManager() *AccountManager {
	var manager = AccountManager{
		Accounts: make(map[string]*Account),
	}

	return &manager
}

func (mgr *AccountManager) InitFromViper() error {
	activeAccounts := viper.GetStringSlice("general.accounts")

	log.S().Infof("Actived accounts: %s", activeAccounts)

	for _, account := range activeAccounts {
		accountInstance, err := NewAccount(account)
		if err != nil {
			return err
		}
		mgr.Accounts[account] = accountInstance
	}

	return nil
}

func (mgr *AccountManager) SyncAllAccounts() error {
	for _, account := range mgr.Accounts {
		go account.Sync()
	}

	return nil
}

func (mgr *AccountManager) Shutdown() error {
	for _, account := range mgr.Accounts {
		account.Shutdown()
	}

	return nil
}

func (mgr *AccountManager) WaitAll() error {
	for _, account := range mgr.Accounts {
		account.Wait()
	}

	return nil
}

type SyncResult struct {
	Err error
}

type RemoteConfig struct {
	Host           string        `mapstructure:"host"`
	Port           uint16        `mapstructure:"port"`
	Username       string        `mapstructure:"username"`
	Password       string        `mapstructure:"password"`
	KeepAlive      time.Duration `mapstructure:"keep_alive"`
	HoldConnection string        `mapstructure:"hold_connection"`
}

type Local struct {
	Type   string `mapstructure:"type"`
	Folder string `mapstructure:"foler"`
}

type Account struct {
	Remote     RemoteConfig
	Local      Local
	Name       string
	MaxAge     int `mapstructure:"max_age"`
	Client     *client.Client
	ctx        context.Context
	ctxCancelF context.CancelFunc
	resultCh   chan SyncResult
}

func NewAccount(name string) (*Account, error) {
	ctx, cancelF := context.WithCancel(context.Background())
	var account = Account{
		Name:       name,
		ctx:        ctx,
		ctxCancelF: cancelF,
		resultCh:   make(chan SyncResult, 1),
	}

	err := viper.Sub(fmt.Sprintf("accounts.%s", name)).Unmarshal(&account)
	if err != nil {
		return nil, fmt.Errorf("configuration file: %v", err)
	}

	return &account, nil
}

func (acc *Account) Sync() error {
	var (
		rsyncer = NewSyncer(acc.Name, acc.Remote)
	)

	go func() {
		constBF := backoff.NewConstantBackOff(time.Second * 10)
		bf := backoff.WithContext(constBF, acc.ctx)

		err := backoff.Retry(func() error {
			err := acc.doSync(rsyncer)
			if err == nil {
				return nil
			}
			switch errors.No(err) {
			case errors.ECtxClosed:
				return backoff.Permanent(err)
			case errors.ELogin:
				return backoff.Permanent(err)
			default:
				log.S().Infof("account sync error, retrying... %v", err)
				return err
			}
		}, bf)

		log.S().Named("Account").Infof("sync exit with err: %v", err)
		acc.resultCh <- SyncResult{Err: err}
	}()

	return nil
}

func (acc *Account) doSync(rsyncer *RemoteSyncer) error {
	var (
		err error
	)

	if err = rsyncer.ConnectWithTLS(acc.ctx); err != nil {
		return err
	}

	if err = rsyncer.Login(acc.ctx); err != nil {
		return err
	}

	if err = rsyncer.Sync(acc.ctx); err != nil {
		rsyncer.Logout()
		return err
	}

	return nil
}

func (acc *Account) Shutdown() error {
	acc.ctxCancelF()
	return nil
}

func (acc *Account) Wait() error {
	result := <-acc.resultCh

	return result.Err
}
