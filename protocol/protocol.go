package protocol

import (
  "bytes"
  "strconv"
  "crypto/tls"
  "crypto/rand"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/worker"
  "github.com/le0pard/go-falcon/storage"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

type env struct {
  *smtpd.BasicEnvelope
}

var (
  SaveMailChan chan *smtpd.BasicEnvelope
)

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
  // send mail to storage workers
  SaveMailChan <- e.BasicEnvelope
  return nil
}

func onNewMail(c smtpd.Connection, from smtpd.MailAddress) (smtpd.Envelope, error) {
  return &env{new(smtpd.BasicEnvelope)}, nil
}

func loadTLSCerts(config *config.Config) (*tls.Config, error) {
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
  // create queue for emails (100 max)
  SaveMailChan = make(chan *smtpd.BasicEnvelope, 100)
  // start parser and storage workers
  worker.StartWorkers(config, SaveMailChan)
  // buffer
  var bufferServer bytes.Buffer
  bufferServer.WriteString(config.Adapter.Host)
  bufferServer.WriteString(":")
  bufferServer.WriteString(strconv.Itoa(config.Adapter.Port))
  // debug info
  log.Debugf("Mail working on %s", bufferServer.String())
  // config database
  db, err := storage.InitDatabase(config)
  if err != nil {
    log.Errorf("Problem with connection to storage: %s", err)
    return
  }
  // config server
  s := &smtpd.Server{
    Addr:      bufferServer.String(),
    OnNewMail: onNewMail,
    ServerConfig: config,
    DBConn: db,
  }
  // tls certs
  if config.Adapter.Tls {
    cert, err := loadTLSCerts(config)
    if err != nil {
      config.Adapter.Tls = false
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