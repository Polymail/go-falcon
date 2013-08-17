package utils

import (
  stdlog "log"
  "flag"
  "strconv"
  "os"
  "os/signal"
  "syscall"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
)

var (
  configFile = flag.String("config", "config.yml", "YAML config for Falcon")
  logFile = flag.String("log", "", "Log file")
  pidFile = flag.String("pid", "", "File for pid file")
  verbose = flag.Bool("v", false, "verbose mode")
  loggerFileDescr  *os.File
  errorFile         error
)

// signals

func listenSignals() {
  go func() {
    sig := make(chan os.Signal, 2)
    signal.Notify(sig, syscall.SIGUSR1)
    defer signal.Stop(sig)
    for {
      <-sig
      if *logFile != "" {
        loggerFileDescr.Close()
        setLoggerOutput()
      }
    }
  }()
}

// setLoggerOutput

func setLoggerOutput() {
  if *logFile != "" {
    loggerFileDescr, errorFile = os.OpenFile(*logFile, os.O_WRONLY | os.O_CREATE, 0640)
    if errorFile != nil {
      log.SetTarget(stdlog.New(os.Stdout, "", stdlog.LstdFlags))
      log.Errorf("Error open file %v", errorFile)
      *logFile = ""
    } else {
      log.SetTarget(stdlog.New(loggerFileDescr, "", stdlog.LstdFlags))
    }
  } else {
    log.SetTarget(stdlog.New(os.Stdout, "", stdlog.LstdFlags))
  }
}

// InitShellParser return config var
func InitDaemon() (*config.Config, error) {
  flag.Parse()
  // signals
  listenSignals()
  // set logger
  setLoggerOutput()
  // info
  log.StartupInfo()
  // config
  log.Infof("Using config file %s", *configFile)
  yamlConfig, err := config.ReadConfig(*configFile)
  if err != nil {
    return nil, err
  }
  // verbose
  yamlConfig.Log.Debug = *verbose
  log.Debug = yamlConfig.Log.Debug
  // write pid
  if *pidFile != "" {
    pid := strconv.Itoa(os.Getpid())
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