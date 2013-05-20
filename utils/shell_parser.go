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

// InitShellParser return config var
func InitShellParser() (*config.Config, error) {
  flag.Parse()
  // config
  log.Infof("Using config file %s", *configFile)
  yamlConfig, err := config.ReadConfig(*configFile)
  if err != nil {
    return nil, err
  }
  // verbose
  yamlConfig.Log.Debug = *verbose
  setLogger(yamlConfig)
  //
  return yamlConfig, nil
}

// setLogger set logger debug mode
func setLogger(config *config.Config) {
  log.Debug = config.Log.Debug
}