package account

import (
	"context"
	"fmt"
	"net"
	"time"

	"goimap/pkg/errors"
	wraplogger "goimap/pkg/log"
	"goimap/pkg/tools"

	"github.com/cenkalti/backoff/v4"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/fatih/color"
	log "go.uber.org/zap"
)

type RemoteSyncer struct {
	config RemoteConfig
	c      *client.Client
	name   string
}

func DoneOrCanceled(ctx context.Context, done <-chan error) (bool, error) {
	select {
	case <-ctx.Done():
		return true, nil
	case err := <-done:
		return false, err
	}
}

func NewSyncer(name string, config RemoteConfig) *RemoteSyncer {
	return &RemoteSyncer{
		config: config,
		c:      nil,
		name:   name,
	}
}

func (s *RemoteSyncer) ConnectWithTLS(ctx context.Context) error {
	var (
		canceled bool
		err      error
		c        *client.Client
	)

	async := tools.RunAsync(func() error {
		dialer := net.Dialer{
			// Time to wait to establish a connection
			Timeout:   time.Second * 10,
			KeepAlive: s.config.KeepAlive * time.Second,
		}

		addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
		if c, err = client.DialWithDialerTLS(&dialer, addr, nil); err != nil {
			return errors.Connect(err)
		}

		return nil
	})

	if canceled, err = async.WaitContext(ctx).DoneOrCanceled(); canceled {
		return errors.CtxClosed()
	}

	if err != nil {
		return err
	}

	c.ErrorLog = wraplogger.NewWrapLogger()
	c.Timeout = time.Second * 60
	log.S().Named("Syncer").Infof("[OK] connect to account %s", s.name)

	s.c = c
	return nil
}

func (s *RemoteSyncer) Login(ctx context.Context) error {
	var (
		canceled bool
		err      error
	)

	log.S().Named("Syncer").Infof("Login %s ...", s.name)

	async := tools.RunAsync(func() error {
		err := s.c.Login(s.config.Username, s.config.Password)
		if err != nil && err != client.ErrAlreadyLoggedIn {
			return err
		}
		return nil
	})

	if canceled, err = async.Wait().DoneOrCanceled(); canceled {
		return errors.CtxClosed()
	}

	if err == nil {
		log.S().Named("Syncer").Infof("[OK] Login %s", s.name)
	}

	return err
}

func (s *RemoteSyncer) Sync(ctx context.Context) error {
	var (
		err      error
		retryCnt = 3
	)

	bf := backoff.WithContext(backoff.NewConstantBackOff(time.Second*2), ctx)

	err = backoff.Retry(func() error {
		if err := s.doSync(ctx); err != nil {
			retryCnt -= 1
			if retryCnt == 0 {
				log.S().Named("Syncer").Infof("[FAIL] max retry times reached: %v", err)
				return backoff.Permanent(err)
			}
			log.S().Named("Syncer").Warnf("[FAIL] retry sync(%d): %v", retryCnt, err)
			return err
		}

		return nil
	}, bf)

	// if backoff is canceled by calling context's cancel function
	// err is ctx.Err()
	if err.Error() == "context canceled" {
		err = errors.CtxClosed()
	}
	return err
}

func (s *RemoteSyncer) doSync(ctx context.Context) error {
	var (
		err       error
		mailboxes = make([]*imap.MailboxInfo, 0)
	)

	log.S().Named("Syncer").Infof("Start syncing: %s", s.name)

	mailboxes, err = s.ListMailboxes(ctx)
	log.S().Named("Syncer").
		Infof("Find mailboxes: %v",
			tools.MapSlice(mailboxes, func(mb *imap.MailboxInfo) string {
				return color.YellowString(mb.Name)
			}),
		)

	for _, mb := range mailboxes {
		s.Select(ctx, mb)
		ids, _ := s.ListMailsByCondition(ctx, mb)
		seqset := new(imap.SeqSet)
		seqset.AddNum(ids...)
		messages := make(chan *imap.Message, 10)
		done := make(chan error, 1)
		go func() {
			done <- s.c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
		}()

		log.S().Infoln("Last 4 messages:")
		for msg := range messages {
			log.S().Infoln("* " + msg.Envelope.Subject)
		}
	}

	log.S().Infoln(err)

	return errors.Foreign(errors.New("qqq"))
}

func (s *RemoteSyncer) Select(ctx context.Context, mb *imap.MailboxInfo) error {
	var (
		mbox     *imap.MailboxStatus
		canceled bool
		err      error
	)

	async := tools.RunAsync(func() error {
		var err error
		mbox, err = s.c.Select(mb.Name, false)
		if err != nil {
			return err
		}
		return nil
	})

	if canceled, err = async.WaitContext(ctx).DoneOrCanceled(); canceled {
		return errors.CtxClosed()
	}

	if err != nil {
		return errors.Select(err)
	}

	log.S().Named("Syncer").Infof("[INFO] Mailbox(%s) contains total %d messages",
		mbox.Name, mbox.Messages)

	return nil
}

func (s *RemoteSyncer) ListMailboxes(ctx context.Context) ([]*imap.MailboxInfo, error) {
	var (
		canceled  bool
		err       error
		mailboxes = make([]*imap.MailboxInfo, 0)
	)

	async := tools.RunAsync(func() error {
		var (
			listDone    = make(chan error, 1)
			mailboxesCh = make(chan *imap.MailboxInfo, 10)
		)

		go func() {
			listDone <- s.c.List("", "*", mailboxesCh)
		}()

		for mb := range mailboxesCh {
			mailboxes = append(mailboxes, mb)
		}

		return <-listDone
	})

	if canceled, err = async.WaitContext(ctx).DoneOrCanceled(); canceled {
		return nil, errors.CtxClosed()
	}

	if err != nil {
		return nil, errors.ListMailbox(err)
	}

	return mailboxes, nil
}

func (s *RemoteSyncer) ListMailsByCondition(ctx context.Context, mbox *imap.MailboxInfo) ([]uint32, error) {
	var (
		canceled bool
		err      error
		ids      = make([]uint32, 0)
	)

	async := tools.RunAsync(func() error {
		criteria := imap.NewSearchCriteria()
		criteria.Since = time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC)

		var err error
		ids, err = s.c.Search(criteria)

		return err
	})

	if canceled, err = async.WaitContext(ctx).DoneOrCanceled(); canceled {
		return nil, errors.CtxClosed()
	}

	if err != nil {
		return ids, errors.ListMail(err)
	}
	log.S().Named("Syncer").Infoln("IDs found:", ids)

	return ids, nil
}

func (s *RemoteSyncer) Logout() {
	s.c.Logout()
}
