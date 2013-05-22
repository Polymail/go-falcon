package storage

import (
  "strings"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/channels"
  "github.com/le0pard/go-falcon/parser"
  "github.com/le0pard/go-falcon/storage/postgresql"
)

type DBConn struct {
  *sql.DB
}

// init database conn
func InitDatabase(config *config.Config) (*DBConn, error) {
  switch strings.ToLower(config.Storage.Adapter) {
  case "postgresql":
    db, err := postgresql.InitDatabase(config)
    return &DBConn{db}, err
  default
    return nil, errors.New("invalid database adapter")
  }
}