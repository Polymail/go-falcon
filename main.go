package main

import (
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/daemon"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol"
  "github.com/le0pard/go-falcon/proxy"

  "fmt"
  "net/http"
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
  // start nginx proxy
  if globalConfig.Proxy.Enabled {
    proxy.StartNginxHTTPProxy(globalConfig)
  } else {
    go func() {
      log.Infof("%+v", http.ListenAndServe(fmt.Sprintf("%s:%d", globalConfig.Adapter.Host, globalConfig.Adapter.Port + 2000), nil))
    }()
  }
  // start pop3 server
  protocol.StartPop3Server(globalConfig)
  // start smtp server
  protocol.StartSmtpServer(globalConfig)
}
