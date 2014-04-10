package main

import (
  stdlog "log"
  "flag"
  "strconv"
  "os"
  "os/signal"
  "syscall"
  "runtime"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/proxy"
  "github.com/le0pard/go-falcon/protocol"
)

var (
  gConfig config.Config

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
  f, err := os.OpenFile(pidFile, os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0666)
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
    loggerFileDescr, errorFile = os.OpenFile(*logFile, os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0640)
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
func initDaemon() (*config.Config, error) {
  flag.Parse()
  // set logger
  setLoggerOutput()
  // signals
  listenSignals()
  // info
  log.StartupInfo()
  // config
  log.Infof("Using config file %s", *configFile)
  yamlConfig, err := config.ReadConfig(*configFile)
  if err != nil {
    return nil, err
  }
  // verbose
  if *verbose == true {
    yamlConfig.Log.Debug = *verbose
  }
  log.Debug = yamlConfig.Log.Debug
  // write pid
  if *pidFile != "" {
    writePidInFile(*pidFile)
  }
  // retrun config
  return yamlConfig, nil
}

func main() {
  // parse shell and config
  gConfig, err := initDaemon()
  if err != nil {
    return
  }
  // conf
  log.Debugf("Loaded config: %v", gConfig)
  // set runtime
  runtime.GOMAXPROCS(gConfig.Daemon.Max_Procs)
  // start nginx proxy
  proxy.StartNginxHTTPProxy(gConfig)
  // start pop3 server
  protocol.StartPop3Server(gConfig)
  // start smtp server
  protocol.StartSmtpServer(gConfig)
}
