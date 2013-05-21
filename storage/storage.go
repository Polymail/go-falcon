package storage

import (
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/channels"
  //"github.com/le0pard/go-falcon/storage/postgresql"
)


func startParserAndStorageWorker(config *config.Config) {
  log.Debugf("Starting storage worker")
  for {
    client := <-channels.SaveMailChan
    log.Debugf("Mail received to storage: %v", client)
  }
}


func StartStorageWorkers(config *config.Config) {
  for i := 0; i < config.Storage.Pool; i++ {
    go startParserAndStorageWorker(config)
  }
}