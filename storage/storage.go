package storage

import (
  "strings"
  "errors"
  "database/sql"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/utils"
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

func (db *DBConn) CheckUser(username, cramPassword, cramSecret string) (int, error) {
  log.Debugf("AUTH by %s / %s", username, cramPassword)
  rows, err := db.DB.Query(db.config.Storage.Mailbox_Sql, username)
  if err != nil {
    log.Errorf("Mailbox SQL error: %v", err)
    return 0, err
  }
  defer rows.Close()
  for rows.Next() {
      var id int
      var password string
      if err := rows.Scan(&id, &password); err != nil {
          log.Errorf("Your mailbox SQL must return 'id' and 'password' fields: %v", err)
          return 0, err
      }
      if !utils.CheckCramMd5Pass(password, cramPassword, cramSecret) {
        log.Errorf("User %s send invalid password", username)
        return 0, errors.New("The user have invalid password")
      }
      return id, nil
  }
  log.Errorf("Coudn't found such user %s with password", username)
  return 0, errors.New("Coudn't found such user")
}