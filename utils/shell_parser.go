package utils

import (
  "flag"
  "github.com/le0pard/go-falcon/log"
  //"launchpad.net/goyaml"
)

var (
  //yamlConfig *yaml.File
  configFile = flag.String("config", "config.yaml", "YAML config for Falcon")
)

func InitShellParser() {
  flag.Parse()
  log.Infof("Using config file %s", *configFile)
  /*
  yamlConfig, err := yaml.ReadFile(*configFile)
  _, err := yaml.ReadFile(*configFile)
  if err != nil {
    log.Errorf("Error read file: readfile(%q): %s", *configFile, err)
  }
  */
}

/*
func GetSmtpHost() (string, error) {
  return yamlConfig.Get("smtp.hostname")
}
*/