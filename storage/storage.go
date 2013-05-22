package storage

import (
  "strings"
  "errors"
  "database/sql"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/storage/postgresql"
)

type DBConn struct {
  DB *sql.DB
  config *config.Config
}

// init database conn
func InitDatabase(config *config.Config) (*DBConn, error) {
  switch strings.ToLower(config.Storage.Adapter) {
  case "postgresql":
    db, err := postgresql.InitDatabase(config)
    return &DBConn{ DB: db, config: config }, err
  default:
    return nil, errors.New("invalid database adapter")
  }
}

func (db *DBConn) CheckUser(username string, password string) {
  log.Debugf("AUTH by %s / %s", username, password)
  rows, err := db.DB.Query("SELECT 1 as name")
  if err != nil {
    log.Errorf("%v", err)
  }
  for rows.Next() {
      var name string
      if err := rows.Scan(&name); err != nil {
          log.Errorf("%v", err)
      }
      log.Debugf("SQL: %s\n", name)
  }
}