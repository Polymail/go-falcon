package utils

import (
  "fmt"
  "bytes"
  "net/http"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
)

var (
  gConfig *config.Config
)

// If running Nginx as a proxy, give Nginx the IP address and port for the SMTP server
// Primary use of Nginx is to terminate TLS so that Go doesn't need to deal with it.
// This could perform auth and load balancing too
// See http://wiki.nginx.org/MailCoreModule
func StartNginxHTTPProxy(config *config.Config) {
  gConfig := config
  if gConfig.Proxy.Enabled == true {
    go nginxHTTPAuth()
  }
}

func nginxHTTPAuth() {
  http.HandleFunc("/", nginxHTTPAuthHandler)
  // listener
  var buffer bytes.Buffer
  buffer.WriteString(gConfig.Proxy.Host)
  buffer.WriteString(":")
  buffer.WriteString(string(gConfig.Proxy.Port))
  err := http.ListenAndServe(buffer.String(), nil)
  if err != nil {
    log.Errorf("ListenAndServe: %v", err)
  }
}

func nginxHTTPAuthHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Auth-Status", "OK")
  w.Header().Add("Auth-Server", gConfig.Adapter.Host)
  w.Header().Add("Auth-Port", string(gConfig.Adapter.Port))
  fmt.Fprint(w, "")
}