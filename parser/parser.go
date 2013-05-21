package parser

import (
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

func ParseMail(mail *smtpd.BasicEnvelope) {
  log.Debugf("Mail received to parser: %v", mail)
}