package channels

import (
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

var (
  SaveMailChan chan *smtpd.BasicEnvelope
)