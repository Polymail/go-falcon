package utils

import (
  "flag"
  "strconv"
  "os"
  "syscall"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
)

var (
  configFile = flag.String("config", "config.yml", "YAML config for Falcon")
  pidFile = flag.String("pid", "", "file for pid file")
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
  // write pid
  if *pidFile != "" {
    pid := strconv.Itoa(syscall.Getpid())
    f, err := os.OpenFile(*pidFile, os.O_RDWR | os.O_CREATE, 0666)
    if err == nil {
      defer f.Close()
      _, err = f.WriteString(pid)
      if err != nil {
        log.Errorf("Error write pid to file: %v", err)
      }
    } else {
      log.Errorf("Open file for pid: %v", err)
    }
  }
  // retrun config
  return yamlConfig, nil
}

// setLogger set logger debug mode
func setLogger(config *config.Config) {
  log.Debug = config.Log.Debug
}