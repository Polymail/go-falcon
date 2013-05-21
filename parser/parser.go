package parser

import (
  "time"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

func ParseMail(mail *smtpd.BasicEnvelope) {
  time.Sleep(1000 * time.Millisecond)
  log.Debugf("Mail received to parser: %v", mail)
}