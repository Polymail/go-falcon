package main

import (
  "runtime"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/utils"
  "github.com/le0pard/go-falcon/proxy"
  "github.com/le0pard/go-falcon/protocol"
)

var (
  gConfig config.Config
)

func main() {
  // parse shell and config
  gConfig, err := utils.InitDaemon()
  if err != nil {
    return
  }
  // conf
  log.Debugf("Loaded config: %v", gConfig)
  // set runtime
  runtime.GOMAXPROCS(gConfig.Daemon.Max_Procs)
  // start nginx proxy
  proxy.StartNginxHTTPProxy(gConfig)
  // start pop3 server
  protocol.StartPop3Server(gConfig)
  // start smtp server
  protocol.StartSmtpServer(gConfig)
}
