package main

import (
  stdlog "log"
  "os"
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
  log.SetTarget(stdlog.New(os.Stdout, "", stdlog.LstdFlags))
  log.StartupInfo()
  // parse shell and config
  gConfig, err := utils.InitShellParser()
  if err != nil {
    return
  }
  // conf
  log.Debugf("Loaded config: %v", gConfig)
  // start nginx proxy
  proxy.StartNginxHTTPProxy(gConfig)
  // start protocol listeners
  protocol.StartMailServer(gConfig)
}
