package main

import (
	"io"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

const (
	// DefaultNetwork network name to be used if none specified
	DefaultNetwork = "freenode"
	// DefaultNetworkURL url to be used if none specified
	DefaultNetworkURL = "irc.freenode.net:7000"
	// DefaultNick placeholder nick
	DefaultNick = "here your nickname"
	// DefaultPassword placeholder pass
	DefaultPassword = "here your password"
	// DefaultIdent placeholder ident
	DefaultIdent = "here your name to the network"
	// DefaultNickserv placeholder nickserv command
	DefaultNickserv = "delete this unless nickserv otherwise add identify password"
	// DefaultChannel placeholder list of channels
	DefaultChannel = "channels separated by comma, dont use hash"

	// KNetworkURL network url key
	KNetworkURL = "networkurl"
	// KNickName bot nickname key
	KNickName = "nickname"
	// KPassword password key
	KPassword = "password"
	// KIdent ident key
	KIdent = "ident"
	// KNickserv nickserv identify key
	KNickserv = "nickserv"
	// KChannel channels key
	KChannel = "channels"
)

// Config holds the config for a bot
type Config struct {
	NetworkName string
	NetworkURL  string
	NickName    string
	Password    string
	Ident       string
	NickservCmd string
	Channels    []string
}

// Write writes the config to the file writer
func (c *Config) Write(file io.Writer) error {
	return writeConfig(file, c)
}

// LoadConfig loads config from an ini file
func LoadConfig(fileName, networkName string) (*Config, error) {
	f, err := ini.Load(fileName)
	if err != nil {
		return nil, errors.Wrap(err, "loading config file")
	}
	nSection := f.Section(networkName)
	if nSection == nil {
		return nil, errors.Errorf("no config for network %q", networkName)
	}
	return &Config{
		NetworkURL:  nSection.Key(KNetworkURL).String(),
		NickName:    nSection.Key(KNickName).String(),
		Password:    nSection.Key(KPassword).String(),
		Ident:       nSection.Key(KIdent).String(),
		NickservCmd: nSection.Key(KNickserv).String(),
		Channels:    nSection.Key(KChannel).Strings(","),
	}, nil
}

// writeConfig writes the passed config or a sample one into the passed file
func writeConfig(file io.Writer, c *Config) error {
	if c == nil {
		c = &Config{
			NetworkName: DefaultNetwork,
			NetworkURL:  DefaultNetworkURL,
			NickName:    DefaultNick,
			Password:    DefaultPassword,
			Ident:       DefaultIdent,
			NickservCmd: DefaultNickserv,
			Channels:    []string{DefaultChannel},
		}
	}
	f := ini.Empty()
	f.NewSections(c.NetworkName)
	nSection := f.Section(c.NetworkName)
	nSection.Comment = "the name you choose for your network, if different from default see -h"
	k, _ := nSection.NewKey(KNetworkURL, c.NetworkURL)
	k.Comment = "please include port, bear in mind that we will use SSL"
	nSection.NewKey(KNickName, c.NickName)
	nSection.NewKey(KPassword, c.Password)
	nSection.NewKey(KIdent, c.Ident)
	nSection.NewKey(KNickserv, c.NickservCmd)
	nSection.NewKey(KChannel, strings.Join(c.Channels, ","))

	_, err := f.WriteTo(file)
	if err != nil {
		return errors.Wrap(err, "writing config file")
	}
	return nil
}