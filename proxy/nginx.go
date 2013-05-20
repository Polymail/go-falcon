package proxy

import (
  "fmt"
  "strconv"
  "bytes"
  "net/http"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
)

// If running Nginx as a proxy, give Nginx the IP address and port for the SMTP server
// Primary use of Nginx is to terminate TLS so that Go doesn't need to deal with it.
// This could perform auth and load balancing too
// See http://wiki.nginx.org/MailCoreModule
func StartNginxHTTPProxy(config *config.Config) {
  if config.Proxy.Enabled == true {
    go nginxHTTPAuth(config)
  }
}

func nginxHTTPAuth(config *config.Config) {
  // handle
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    nginxHTTPAuthHandler(w, r, config)
  })
  // listener
  var buffer bytes.Buffer
  buffer.WriteString(config.Proxy.Host)
  buffer.WriteString(":")
  buffer.WriteString(strconv.Itoa(config.Proxy.Port))
  //
  log.Debugf("Nginx proxy working on %s", buffer.String())
  //
  err := http.ListenAndServe(buffer.String(), nil)
  if err != nil {
    log.Errorf("Nginx proxy: %v", err)
  }
}

func nginxHTTPAuthHandler(w http.ResponseWriter, r *http.Request, config *config.Config) {
  log.Debugf("Nginx proxy get request: %v", r)
  //
  w.Header().Add("Auth-Status", "OK")
  w.Header().Add("Auth-Server", config.Adapter.Host)
  w.Header().Add("Auth-Port", strconv.Itoa(config.Adapter.Port))
  fmt.Fprint(w, "")
}