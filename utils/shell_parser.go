package utils

import (
  "flag"
  "os"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
)

var (
  configFile = flag.String("config", "config.yaml", "YAML config for Falcon")
)

func InitShellParser() {
  flag.Parse()
  log.Infof("Using config file %s", *configFile)
  yamlConfig, err := config.ReadConfig(*configFile)
  if err != nil {
    os.Exit(1)
  }
  log.Noticef("--- t:\n%v\n\n", yamlConfig)
}