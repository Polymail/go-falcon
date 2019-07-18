package main

import (
	"github.com/Polymail/go-falcon/config"
	"github.com/Polymail/go-falcon/daemon"
	"github.com/Polymail/go-falcon/log"
	"github.com/Polymail/go-falcon/protocol"
	"github.com/Polymail/go-falcon/proxy"
)

var (
	globalConfig config.Config
)

func main() {
	// parse shell and config
	globalConfig, err := daemon.InitDaemon()
	if err != nil {
		return
	}
	// conf
	log.Debugf("Loaded config: %+v", globalConfig)
	// start nginx proxy
	proxy.StartNginxHTTPProxy(globalConfig)
	// start pop3 server
	protocol.StartPop3Server(globalConfig)
	// start smtp server
	protocol.StartSmtpServer(globalConfig)
}
