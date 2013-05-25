package parser

import (
  "bytes"
  "net/mail"
  "io/ioutil"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

type EmailParser struct {
  env *smtpd.BasicEnvelope
}

func (parser *EmailParser) ParseMail(env *smtpd.BasicEnvelope) {
  parser.env = env
  msg, err := mail.ReadMessage(bytes.NewBuffer(parser.env.MailBody))
  if err != nil {
    log.Errorf("Failed parsing message: %v", err)
    return
  }
  log.Debugf("headers: %v", msg.Header)
  body, err := ioutil.ReadAll(msg.Body)
  if err != nil {
    log.Errorf("Failed parsing message: %v", err)
    return
  }
  log.Debugf("body: %v", string(body))
}