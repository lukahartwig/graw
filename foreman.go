package graw

import (
	"log"

	"github.com/lukahartwig/graw/botfaces"
	"github.com/lukahartwig/graw/reddit"
)

func launch(
	handler interface{},
	kill chan bool,
	errs <-chan error,
	logger *log.Logger,
) (
	func(),
	func() error,
	error,
) {
	if setup, ok := handler.(botfaces.Loader); ok {
		if err := setup.SetUp(); err != nil {
			return nil, nil, err
		}
	}

	tear := func() {
		if tear, ok := handler.(botfaces.Tearer); ok {
			tear.TearDown()
		}
	}

	foremanKiller := make(chan bool)
	foremanError := make(chan error)

	go func() {
		foremanError <- foreman(foremanKiller, kill, errs, logger)
	}()

	stop := func() {
		defer tear()
		close(foremanKiller)
	}

	wait := func() error {
		defer tear()
		return <-foremanError
	}

	return stop, wait, nil
}

func foreman(
	kill <-chan bool,
	killChildren chan<- bool,
	errs <-chan error,
	logger *log.Logger,
) error {
	defer close(killChildren)
	for {
		select {
		case <-kill:
			return nil
		case err := <-errs:
			switch err {
			case nil:
			case reddit.ErrBusy:
				logger.Printf("Reddit was busy; staying up.")
			case reddit.ErrGateway:
				logger.Printf("Bad gateway error; staying up.")
			case reddit.ErrGatewayTimeout:
				logger.Printf("Gateway timeout; staying up.")
			default:
				return err
			}
		}
	}
}
