package daemon

import (
  "flag"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/log"
  stdlog "log"
  "os"
  "os/signal"
  "strconv"
  "syscall"
)

var (
  configFile = flag.String("config", "config.yml", "YAML config for Falcon")
  logFile    = flag.String("log", "", "Log file")
  pidFile    = flag.String("pid", "", "File for pid file")
  verbose    = flag.Bool("V", false, "Verbose mode")

  loggerFileDescr *os.File
  errorFile       error
)

// signals

func listenSignals() {
  go func() {
    signals := make(chan os.Signal, 2)
    signal.Notify(signals, syscall.SIGUSR1, syscall.SIGUSR2)
    defer signal.Stop(signals)
    for {
      sig := <-signals
      switch sig {
      case syscall.SIGUSR1:
        if *logFile != "" {
          loggerFileDescr.Close()
          setLoggerOutput()
        }
      case syscall.SIGUSR2:
        // TODO: reload config
      }
    }
  }()
}

// write pid in file

func writePidInFile(pidFile string) {
  pid := strconv.Itoa(os.Getpid())
  f, err := os.OpenFile(pidFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
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

// setLoggerOutput

func setLoggerOutput() {
  if *logFile != "" {
    loggerFileDescr, errorFile = os.OpenFile(*logFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0640)
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
  // set logger
  setLoggerOutput()
  // signals
  listenSignals()
  // info
  log.StartupInfo()
  // config
  log.Infof("Using config file %s", *configFile)
  globalConfig, err := config.ReadConfig(*configFile)
  if err != nil {
    return nil, err
  }
  // verbose
  if *verbose == true {
    globalConfig.Log.Debug = *verbose
  }
  log.Debug = globalConfig.Log.Debug
  // write pid
  if *pidFile != "" {
    writePidInFile(*pidFile)
  }
  // retrun config
  return globalConfig, nil
}
