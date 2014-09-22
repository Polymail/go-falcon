package main

import (
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/daemon"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol"
  "github.com/le0pard/go-falcon/proxy"
  _ "net/http/pprof"
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
  // set runtime, disabled
  //runtime.GOMAXPROCS(globalConfig.Daemon.Max_Procs)
  // start nginx proxy
  proxy.StartNginxHTTPProxy(globalConfig)
  // start pop3 server
  protocol.StartPop3Server(globalConfig)
  // start smtp server
  protocol.StartSmtpServer(globalConfig)
}
