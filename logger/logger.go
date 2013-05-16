package logger

import (
  "fmt"
  "os"
  "log"
  "runtime"
)

func Info(a ...interface{}) {
  fmt.Fprintln(os.Stdout, a...)
}

func FatalError(format string, v ...interface{}) {
  log.Fatalf(format, v...)
}

func StartupInfo() {
  Info("HamstersHorde, built with Go", runtime.Version())
  Info("(c) leopard aka Alexey Vasiliev")
}