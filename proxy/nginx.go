package proxy

import (
  "fmt"
  "strconv"
  "bytes"
  "net/http"
  "strings"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/storage"
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

// nginx auth server

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

// nginx auth by nginx headers

func nginxHTTPAuthHandler(w http.ResponseWriter, r *http.Request, config *config.Config) {
  log.Debugf("Nginx proxy get request: %v", r)
  
  protocol := r.Header.Get("Auth-Protocol")

  if strings.ToLower(protocol) == "smtp" {
    authMethod := r.Header.Get("Auth-Method")
    username := r.Header.Get("Auth-User")
    password := r.Header.Get("Auth-Pass")
    secret := ""
    if strings.ToLower(authMethod) == "cram-md5" {
      secret = r.Header.Get("Auth-Salt")
    }
    // db connect
    db, err := storage.InitDatabase(config)
    if err != nil {
      log.Errorf("Couldn't connect to database: %v", err)
      nginxResponseFail(w, r)
      return
    }
    defer db.Close()
    db.DB.SetMaxIdleConns(-1)
    id, err := db.CheckUser(username, password, secret)
    if err != nil {
      nginxResponseFail(w, r)
      return
    }
    nginxResponseSuccess(config, w, strconv.Itoa(id))
  } else {
    nginxResponseFail(w, r)
  }
}

// success auth response

func nginxResponseSuccess(config *config.Config, w http.ResponseWriter, userId string) {
  w.Header().Add("Auth-Status", "OK")
  w.Header().Add("Auth-Server", config.Adapter.Host)
  w.Header().Add("Auth-Port", strconv.Itoa(config.Adapter.Port))
  // return mailbox id
  if userId != "" {
    w.Header().Add("Auth-User", userId)
  }
  // empty body
  fmt.Fprint(w, "")
}

// fail auth response

func nginxResponseFail(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Auth-Status", "Invalid login or password")
  // login attempt
  loginAttempt := r.Header.Get("Auth-Login-Attempt")
  if loginAttempt != "" {
    loginAttemptInt, err := strconv.Atoi(loginAttempt)
    if err == nil {
      if loginAttemptInt < 10 {
        w.Header().Add("Auth-Wait", "3")
      }
    }
  }
  // empty body
  fmt.Fprint(w, "")
}
