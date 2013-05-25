package parser

import (
  "bytes"
  "net/mail"
  "io/ioutil"
  "time"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

type ParsedEmail struct {
  env *smtpd.BasicEnvelope

  Subject       string
  Date          time.Time
  From          mail.Address
  To            mail.Address
  Headers       mail.Header
  EmailBody     []byte
}

// parse headers

func (email *ParsedEmail) parseEmailHeaders(msg *mail.Message) {
  var emailHeader = ""
  var err error

  email.Headers = msg.Header
  email.Subject = email.Headers.Get("Subject")
  email.Date, err = msg.Header.Date()
  if err != nil {
    log.Errorf("Failed parsing date: %v", err)
    email.Date = time.Now()
  }
  // from
  emailHeader = email.Headers.Get("From")
  if emailHeader != "" {
    fromEmail, err := mail.ParseAddress(emailHeader)
    if err != nil {
      email.From = mail.Address{ Address: email.env.From.Email() }
    } else {
      email.From = *fromEmail
    }
  }
  // to
  emailHeader = email.Headers.Get("To")
  if emailHeader != "" {
    toEmail, err := mail.ParseAddress(emailHeader)
    if err != nil {
      email.To = mail.Address{}
    } else {
      email.To = *toEmail
    }
  }
}

// obj

type EmailParser struct {
}

// parse email

func (parser *EmailParser) ParseMail(env *smtpd.BasicEnvelope) {
  email := ParsedEmail{ env: env }
  msg, err := mail.ReadMessage(bytes.NewBuffer(email.env.MailBody))
  if err != nil {
    log.Errorf("Failed parsing message: %v", err)
    return
  }
  body, err := ioutil.ReadAll(msg.Body)
  if err != nil {
    log.Errorf("Failed parsing message: %v", err)
    return
  }
  email.parseEmailHeaders(msg)
  email.EmailBody = body
  log.Debugf("parsed: %v", email)
}