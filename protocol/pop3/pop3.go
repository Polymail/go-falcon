// Package smtpd implements an pop3 server. Hooks are provided to customize
// its behavior.
package pop3

import (
  "bufio"
  "bytes"
  "errors"
  "regexp"
  "fmt"
  "net"
  "os/exec"
  "crypto/tls"
  "strconv"
  "strings"
  "time"
  "unicode"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/storage"
  "github.com/le0pard/go-falcon/utils"
)


// Server is an POP3 server.
type Server struct {
  Addr         string // TCP address to listen on, ":2525" if empty
  Hostname     string // optional Hostname to announce; "" to use system hostname
  ReadTimeout  time.Duration  // optional read timeout
  WriteTimeout time.Duration  // optional write timeout

  TLSconfig *tls.Config // tls config

  ServerConfig *config.Config
  DBConn       *storage.DBConn
}