package protocol

import (
  "bytes"
  "strconv"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

type env struct {
  *smtpd.BasicEnvelope
}

func (e *env) AddRecipient(rcpt smtpd.MailAddress) error {
  // filter for recipient
  /*
  if strings.HasPrefix(rcpt.Email(), "bad@") {
    return errors.New("we don't send email to bad@")
  }
  */
  return e.BasicEnvelope.AddRecipient(rcpt)
}

func (e *env) Close() error {
  log.Infof("Message finished")
  log.Infof("Mail: %v", e.BasicEnvelope)
  return nil
}

func onNewMail(c smtpd.Connection, from smtpd.MailAddress) (smtpd.Envelope, error) {
  log.Infof("new mail from %q", from)
  return &env{new(smtpd.BasicEnvelope)}, nil
}


func StartMailServer(config *config.Config) {
  var buffer bytes.Buffer
  buffer.WriteString(config.Adapter.Host)
  buffer.WriteString(":")
  buffer.WriteString(strconv.Itoa(config.Adapter.Port))
  //
  log.Debugf("Mail working on %s", buffer.String())
  //
  s := &smtpd.Server{
    Addr:      buffer.String(),
    OnNewMail: onNewMail,
    ServerConfig: *config,
  }
  error := s.ListenAndServe()
  if error != nil {
    log.Errorf("Mail server: %v", error)
  }
}