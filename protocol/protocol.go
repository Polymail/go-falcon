package protocol

import (
  "crypto/tls"
  "crypto/rand"
  "fmt"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/worker"
  "github.com/le0pard/go-falcon/storage"
  "github.com/le0pard/go-falcon/protocol/pop3"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

const (
  EMAIL_CHANNEL_SIZE = 20
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

// load POP3 TLS certs

func loadPop3TLSCerts(config *config.Config) (*tls.Config, error) {
  cert, err := tls.LoadX509KeyPair(config.Pop3.Ssl_Pub_Key, config.Pop3.Ssl_Prv_Key)
  if err != nil {
    log.Errorf("POP3: There was a problem with loading the certificate: %s", err)
    return nil, err
  }
  TLSconfig := &tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.VerifyClientCertIfGiven, ServerName: config.Pop3.Ssl_Hostname, InsecureSkipVerify: true}
  TLSconfig.Rand = rand.Reader
  return TLSconfig, nil
}

// load SMTP TLS certs

func loadSmtpTLSCerts(config *config.Config) (*tls.Config, error) {
  cert, err := tls.LoadX509KeyPair(config.Adapter.Ssl_Pub_Key, config.Adapter.Ssl_Prv_Key)
  if err != nil {
    log.Errorf("SMTPD: There was a problem with loading the certificate: %s", err)
    return nil, err
  }
  TLSconfig := &tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.VerifyClientCertIfGiven, ServerName: config.Adapter.Ssl_Hostname, InsecureSkipVerify: true}
  TLSconfig.Rand = rand.Reader
  return TLSconfig, nil
}

// start pop3 in goroot

func goPop3Server(config *config.Config) {
  // server ip:port
  serverBind := fmt.Sprintf("%s:%d", config.Pop3.Host, config.Pop3.Port)
  // debug info
  log.Debugf("POP3 working on %s", serverBind)
  // config database
  db, err := storage.InitDatabase(config)
  if err != nil {
    log.Errorf("Problem with connection to storage: %s", err)
    return
  }
  db.DB.SetMaxIdleConns(2)
  // config server
  s := &pop3.Server{
    Addr:         serverBind,
    Hostname:     config.Pop3.Hostname,
    ServerConfig: config,
    DBConn:       db,
  }
  // tls certs
  if config.Pop3.Tls {
    cert, err := loadPop3TLSCerts(config)
    if err != nil {
      config.Pop3.Tls = false
    } else {
      s.TLSconfig = cert
    }
  }
  // server
  error := s.ListenAndServe()
  if error != nil {
    log.Errorf("POP3 server: %v", error)
  }
}

// start pop3 server

func StartPop3Server(config *config.Config) {
  if config.Pop3.Enabled {
    go goPop3Server(config)
  }
}

// start smtp server

func StartSmtpServer(config *config.Config) {
  // create queue for emails
  SaveMailChan = make(chan *smtpd.BasicEnvelope, EMAIL_CHANNEL_SIZE)
  // start parser and storage workers
  worker.StartWorkers(config, SaveMailChan)
  // server ip:port
  serverBind := fmt.Sprintf("%s:%d", config.Adapter.Host, config.Adapter.Port)
  // debug info
  log.Debugf("SMPTD working on %s", serverBind)
  // config database
  db, err := storage.InitDatabase(config)
  if err != nil {
    log.Errorf("Problem with connection to storage: %s", err)
    return
  }
  db.DB.SetMaxIdleConns(2)
  // config server
  s := &smtpd.Server{
    Addr:         serverBind,
    Hostname:     config.Adapter.Hostname,
    OnNewMail:    onNewMail,
    ServerConfig: config,
    DBConn:       db,
  }
  // tls certs
  if config.Adapter.Tls {
    cert, err := loadSmtpTLSCerts(config)
    if err != nil {
      config.Adapter.Tls = false
    } else {
      s.TLSconfig = cert
    }
  }
  // server
  error := s.ListenAndServe()
  if error != nil {
    log.Errorf("SMPTD server: %v", error)
  }
}