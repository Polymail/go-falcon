package utils

import (
  "flag"
  "github.com/le0pard/go-falcon/logger"
)

var (
  configFile = flag.String("config", "config.yaml", "YAML config")
)

func initShellParser() {
  flag.Parse()
  logger.Info("Using config file", *configFile)
}
