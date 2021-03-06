package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"path/filepath"

	irc "github.com/fluffle/goirc/client"
	"github.com/perrito666/lessisbetterbot/skills"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// bot is the basic irc bot actor.
type bot struct {
	logger *log.Logger
	cfg    *Config
	db     *bolt.DB
}

// live tarts the bot loop.
func (b *bot) live() error {
	cfg := irc.NewConfig(b.cfg.NickName, b.cfg.NickName, b.cfg.Ident)
	cfg.SSL = true
	// ok, yes, thisis harcoded
	cfg.SSLConfig = &tls.Config{ServerName: "irc.freenode.net"}
	cfg.Server = b.cfg.NetworkURL
	cfg.NewNick = func(n string) string { return n + "'" }
	cfg.Pass = b.cfg.Password
	c := irc.Client(cfg)

	// where all the trash goes
	attic, err := bolt.Open(filepath.Join(b.cfg.StorageFolder, "attic.db"), 0600, nil)
	if err != nil {
		return errors.Wrap(err, "opening webpeek database")
	}
	b.db = attic
	defer attic.Close()

	// Create buckets for all skills
	err = attic.Update(func(tx *bolt.Tx) error {
		for _, bucket := range []string{
			"webpeek", skills.USD, skills.REAL,
		} {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "ensuring bucket existence")
	}

	// Add handlers to do things here!
	// e.g. join a channel on connect.
	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			b.logger.Println("connected to freenode")
			if b.cfg.NickservCmd != "" {
				conn.Privmsg("nickserv", fmt.Sprintf("identify %s", b.cfg.NickservCmd))
			}
			for _, channel := range b.cfg.Channels {
				conn.Join("#" + channel)
				b.logger.Printf("joined #%s\n", channel)
			}
		})
	// And a signal on disconnect
	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			b.logger.Println("disconnected from freenode")
			quit <- true
		})

	// do all the handling
	c.HandleFunc(irc.PRIVMSG, b.handleMsg)

	// Tell client to connect.
	b.logger.Println("will connect")
	if err := c.Connect(); err != nil {
		return errors.Wrap(err, "connecting to freenode")
	}
	b.logger.Println("did not fail to connect (?)")

	// Wait for disconnect
	<-quit
	return nil
}
