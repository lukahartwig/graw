package graw

import (
	"fmt"

	"github.com/lukahartwig/graw/botfaces"
	"github.com/lukahartwig/graw/reddit"
	"github.com/lukahartwig/graw/streams"
)

var (
	errPostHandler = fmt.Errorf(
		"you must implement PostHandler to handle subreddit feeds",
	)
	errCommentHandler = fmt.Errorf(
		"you must implement CommentHandler to handle subreddit " +
			"comment feeds",
	)
	errUserHandler = fmt.Errorf(
		"you must implement UserHandler to handle user feeds",
	)
	errLoggedOut = fmt.Errorf(
		"you must be running as a logged in bot to get inbox feeds",
	)
)

// Scan connects any requested logged-out event sources to the given handler,
// making requests with the given script handle. It launches a goroutine for the
// scan. It returns two functions: a stop() function to stop the scan at any
// time, and a wait() function to block until the scan fails.
func Scan(handler interface{}, script reddit.Script, cfg Config) (
	func(),
	func() error,
	error,
) {
	kill := make(chan bool)
	errs := make(chan error)

	if cfg.PostReplies || cfg.CommentReplies || cfg.Mentions || cfg.Messages {
		return nil, nil, errLoggedOut
	}

	if err := connectScanStreams(
		handler,
		script,
		cfg,
		kill,
		errs,
	); err != nil {
		return nil, nil, err
	}

	return launch(handler, kill, errs, logger(cfg.Logger))
}

// connectScanStreams connects the streams a scanner can subscribe to to the
// handler.
func connectScanStreams(
	handler interface{},
	sc reddit.Scanner,
	c Config,
	kill <-chan bool,
	errs chan<- error,
) error {
	if len(c.Subreddits) > 0 {
		ph, ok := handler.(botfaces.PostHandler)
		if !ok {
			return errPostHandler
		}

		if posts, err := streams.Subreddits(
			sc,
			kill,
			errs,
			c.Subreddits...,
		); err != nil {
			return err
		} else {
			go func() {
				for p := range posts {
					errs <- ph.Post(p)
				}
			}()
		}
	}

	if len(c.CustomFeeds) > 0 {
		ph, ok := handler.(botfaces.PostHandler)
		if !ok {
			return errPostHandler
		}

		for user, feeds := range c.CustomFeeds {
			if posts, err := streams.CustomFeeds(
				sc,
				kill,
				errs,
				user,
				feeds...,
			); err != nil {
				return err
			} else {
				go func() {
					for p := range posts {
						errs <- ph.Post(p)
					}
				}()
			}
		}
	}

	if len(c.SubredditComments) > 0 {
		ch, ok := handler.(botfaces.CommentHandler)
		if !ok {
			return errCommentHandler
		}

		if comments, err := streams.SubredditComments(
			sc,
			kill,
			errs,
			c.SubredditComments...,
		); err != nil {
			return err
		} else {
			go func() {
				for c := range comments {
					errs <- ch.Comment(c)
				}
			}()
		}
	}

	if len(c.Users) > 0 {
		uh, ok := handler.(botfaces.UserHandler)
		if !ok {
			return errUserHandler
		}

		for _, user := range c.Users {
			if posts, comments, err := streams.User(
				sc,
				kill,
				errs,
				user,
			); err != nil {
				return err
			} else {
				go func() {
					for p := range posts {
						errs <- uh.UserPost(p)
					}
				}()
				go func() {
					for c := range comments {
						errs <- uh.UserComment(c)
					}
				}()
			}
		}
	}

	return nil
}
