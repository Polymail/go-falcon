package smtpd

// https://github.com/bradfitz/go-smtpd/blob/master/smtpd/smtpd.go
// https://github.com/jda/go-lmtpd/blob/master/lmtpd/lmtpd.go

import (
  "regexp"
)

var (
  rcptToRE = regexp.MustCompile(`[Tt][Oo]:<(.+)>`)
  mailFromRE = regexp.MustCompile(`[Ff][Rr][Oo][Mm]:<(.*)>`)
)