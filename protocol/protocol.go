package protocol

import (
  "bytes"
  "strconv"
  "crypto/tls"
  "crypto/rand"
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
  log.Debugf("Mail received: %v", e.BasicEnvelope)
  return nil
}

func onNewMail(c smtpd.Connection, from smtpd.MailAddress) (smtpd.Envelope, error) {
  return &env{new(smtpd.BasicEnvelope)}, nil
}

func loadTSLCerts(config *config.Config) (*tls.Config, error) {
  cert, err := tls.LoadX509KeyPair(config.Adapter.Ssl_Pub_Key, config.Adapter.Ssl_Prv_Key)
  if err != nil {
    log.Errorf("There was a problem with loading the certificate: %s", err)
    return nil, err
  }
  TLSconfig := &tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.VerifyClientCertIfGiven, ServerName: config.Adapter.Ssl_Hostname}
  TLSconfig.Rand = rand.Reader
  return TLSconfig, nil
}


func StartMailServer(config *config.Config) {
  var buffer bytes.Buffer
  buffer.WriteString(config.Adapter.Host)
  buffer.WriteString(":")
  buffer.WriteString(strconv.Itoa(config.Adapter.Port))
  //
  log.Debugf("Mail working on %s", buffer.String())
  // config server
  s := &smtpd.Server{
    Addr:      buffer.String(),
    OnNewMail: onNewMail,
    ServerConfig: config,
  }
  // certs
  if config.Adapter.Tsl {
    cert, err := loadTSLCerts(config)
    if err != nil {
      config.Adapter.Tsl = false
    } else {
      s.TLSconfig = cert
    }
  }
  // server
  error := s.ListenAndServe()
  if error != nil {
    log.Errorf("Mail server: %v", error)
  }
}