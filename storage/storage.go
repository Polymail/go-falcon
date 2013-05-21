package storage

import (
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/channels"
  "github.com/le0pard/go-falcon/parser"
  //"github.com/le0pard/go-falcon/storage/postgresql"
)


func startParserAndStorageWorker(config *config.Config) {
  log.Debugf("Starting storage worker")
  for {
    envelop := <-channels.SaveMailChan
    parser.ParseMail(envelop)
  }
}


func StartStorageWorkers(config *config.Config) {
  for i := 0; i < config.Storage.Pool; i++ {
    go startParserAndStorageWorker(config)
  }
}