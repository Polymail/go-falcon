package worker

import (
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/parser"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

// start worker
func startParserAndStorageWorker(config *config.Config, channel chan *smtpd.BasicEnvelope) {
  log.Debugf("Starting storage worker")
  for {
    envelop := <- channel
    parser.ParseMail(envelop)
    //postgresql.StoreMail()
  }
}

// workers
func StartWorkers(config *config.Config, channel chan *smtpd.BasicEnvelope) {
  for i := 0; i < config.Storage.Pool; i++ {
    go startParserAndStorageWorker(config, channel)
  }
}