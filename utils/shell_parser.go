package utils

import (
  "flag"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
)

var (
  configFile = flag.String("config", "config.yml", "YAML config for Falcon")
  verbose = flag.Bool("v", false, "verbose mode")
)

func InitShellParser() (*config.Config, error) {
  flag.Parse()
  log.Infof("Using config file %s", *configFile)
  yamlConfig, err := config.ReadConfig(*configFile)
  if err != nil {
    return nil, err
  }
  return yamlConfig, nil
}