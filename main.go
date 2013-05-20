package main

import (
  stdlog "log"
  "os"
  "errors"
  "strings"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/utils"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

var (
  gConfig config.Config
)

type env struct {
  *smtpd.BasicEnvelope
}

func (e *env) AddRecipient(rcpt smtpd.MailAddress) error {
  if strings.HasPrefix(rcpt.Email(), "bad@") {
    return errors.New("we don't send email to bad@")
  }
  return e.BasicEnvelope.AddRecipient(rcpt)
}

func onNewMail(c smtpd.Connection, from smtpd.MailAddress) (smtpd.Envelope, error) {
  log.Infof("ajas: new mail from %q", from)
  return &env{new(smtpd.BasicEnvelope)}, nil
}

func main() {
  log.SetTarget(stdlog.New(os.Stdout, "", stdlog.LstdFlags))
  log.StartupInfo()
  // parse shell and config
  gConfig, err := utils.InitShellParser()
  if err != nil {
    return
  }
  // begin work
  log.Noticef("\n%v\n\n", gConfig)
  // start nginx proxy
  utils.StartNginxHTTPProxy(gConfig)
  // start protocol listeners
  s := &smtpd.Server{
    Addr:      "localhost:2526",
    OnNewMail: onNewMail,
  }
  error := s.ListenAndServe()
  if error != nil {
    log.Infof("ListenAndServe: %v", error)
  }
}
