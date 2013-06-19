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
  if config.Proxy.Enabled {
    go nginxHTTPAuth(config)
  }
}

/*

2013/06/19 13:06:43 DEBUG Command from client QUIT
2013/06/19 13:07:09 DEBUG Nginx proxy get request: &{GET / HTTP/1.0 1 0 map[Auth-Method:[plain] Auth-User:[123] Auth-Pass:[123123] Auth-Protocol:[smtp] Auth-Login-Attempt:[1] Client-Ip:[213.160.145.74] Client-Host:[[UNAVAILABLE]]] 0xc20011a9c0 0 [] false localhost map[] map[] <nil> map[] 127.0.0.1:53266 / <nil>}
2013/06/19 13:07:09 DEBUG Command from client EHLO falcon.rw.rw
2013/06/19 13:07:09 DEBUG Command from client MAIL FROM:<me@fromdomain.com>
2013/06/19 13:07:09 DEBUG mail from: "me@fromdomain.com"
2013/06/19 13:07:09 DEBUG Command from client RCPT TO:<test@todomain.com>
2013/06/19 13:07:09 DEBUG Command from client RCPT TO:<test2@todomain.com>
2013/06/19 13:07:09 DEBUG Command from client RCPT TO:<test3@todomain.com>
2013/06/19 13:07:09 DEBUG Command from client QUIT
2013/06/19 13:07:57 DEBUG Nginx proxy get request: &{GET / HTTP/1.0 1 0 map[Auth-Method:[plain] Auth-User:[123] Auth-Pass:[123123] Auth-Protocol:[smtp] Auth-Login-Attempt:[1] Client-Ip:[213.160.145.74] Client-Host:[[UNAVAILABLE]]] 0xc20011a040 0 [] false localhost map[] map[] <nil> map[] 127.0.0.1:53270 / <nil>}
2013/06/19 13:07:57 DEBUG Command from client EHLO falcon.rw.rw
2013/06/19 13:07:58 DEBUG Command from client MAIL FROM:<me@fromdomain.com>
2013/06/19 13:07:58 DEBUG mail from: "me@fromdomain.com"
2013/06/19 13:07:58 DEBUG Command from client RCPT TO:<test@todomain.com>
2013/06/19 13:07:58 DEBUG Command from client RCPT TO:<test2@todomain.com>
2013/06/19 13:07:58 DEBUG Command from client RCPT TO:<test3@todomain.com>
2013/06/19 13:07:58 DEBUG Command from client QUIT
2013/06/19 13:08:15 DEBUG Nginx proxy get request: &{GET / HTTP/1.0 1 0 map[Auth-Pass:[a6da18f4b2a1b68556fded166007be61] Auth-Protocol:[smtp] Auth-Login-Attempt:[1] Client-Ip:[213.160.145.74] Auth-Method:[cram-md5] Auth-User:[123] Auth-Salt:[<613169054.1371647295@falcon.rw.rw>] Client-Host:[[UNAVAILABLE]]] 0xc20011a880 0 [] false localhost map[] map[] <nil> map[] 127.0.0.1:53272 / <nil>}


HTTP/1.0 200 OK      # this line is actually ignored and may not exist at all
Auth-Status: Invalid login or password
Auth-Wait: 3         # nginx will wait 3 seconds before reading
# client's login/passwd again

2013/06/19 15:22:29 DEBUG XCLIENT info: XCLIENT ADDR=213.160.145.74 LOGIN=test NAME=[UNAVAILABLE]

*/

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