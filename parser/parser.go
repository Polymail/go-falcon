package parser

import (
  "bytes"
  "net/mail"
  "io/ioutil"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

func ParseMail(envelope *smtpd.BasicEnvelope) {
  log.Debugf("Mail received to parser: %v", envelope)
  msg, err := mail.ReadMessage(bytes.NewBuffer(envelope.MailBody))
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