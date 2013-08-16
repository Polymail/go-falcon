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
  "github.com/le0pard/go-falcon/utils"
)

const (
  MAX_AUTH_RETRY = 10
  INVALID_AUTH_WAIT_TIME = "3"
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
  //log.Debugf("Nginx proxy get request: %v", r)

  protocol := strings.ToLower(r.Header.Get("Auth-Protocol"))

  if protocol == "smtp" || protocol == "pop3" {
    authMethod := strings.ToLower(r.Header.Get("Auth-Method"))
    username := r.Header.Get("Auth-User")
    password := r.Header.Get("Auth-Pass")
    secret := ""
    if authMethod == utils.AUTH_CRAM_MD5 || authMethod == utils.AUTH_APOP {
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
    id, pass, err := db.CheckUserWithPass(authMethod, username, password, secret)
    if err != nil {
      nginxResponseFail(w, r)
      return
    }
    nginxResponseSuccess(config, w, protocol, strconv.Itoa(id), pass)
  } else {
    nginxResponseFail(w, r)
  }
}

// success auth response

func nginxResponseSuccess(config *config.Config, w http.ResponseWriter, protocol, userId, password string) {
  w.Header().Add("Auth-Status", "OK")
  if protocol == "smtp" {
    w.Header().Add("Auth-Server", config.Adapter.Host)
    w.Header().Add("Auth-Port", strconv.Itoa(config.Adapter.Port))
    // return mailbox id
    w.Header().Add("Auth-User", userId)
  } else if protocol == "pop3" {
    w.Header().Add("Auth-Server", config.Pop3.Host)
    w.Header().Add("Auth-Port", strconv.Itoa(config.Pop3.Port))
    // return password
    w.Header().Add("Auth-Pass", password)
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
      if loginAttemptInt < MAX_AUTH_RETRY {
        w.Header().Add("Auth-Wait", INVALID_AUTH_WAIT_TIME)
      }
    }
  }
  // empty body
  fmt.Fprint(w, "")
}
