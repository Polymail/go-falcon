package storage

import (
  "strings"
  "errors"
  "database/sql"
  "time"
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

// check username login

func (db *DBConn) CheckUser(username, cramPassword, cramSecret string) (int, error) {
  var (
    id int
    password string
  )
  log.Debugf("AUTH by %s / %s", username, cramPassword)
  err := db.DB.QueryRow(db.config.Storage.Mailbox_Sql, username).Scan(&id, &password)
  if err != nil {
    log.Errorf("User %s doesn't found (sql should return 'id' and 'password' fields): %v", username, err)
    return 0, err
  }
  if !utils.CheckSMTPAuthPass(password, cramPassword, cramSecret) {
    log.Errorf("User %s send invalid password", username)
    return 0, errors.New("The user have invalid password")
  }
  return id, nil
}

// save email and attachments

func (db *DBConn) StoreMail(mailboxId int, subject string, date time.Time, from, from_name, to, to_name, html, text string, rawEmail []byte) (int, error) {
  var (
    id int
  )
  err := db.DB.QueryRow(db.config.Storage.Messages_Sql,
    mailboxId,
    subject,
    date,
    from,
    from_name,
    to,
    to_name,
    html,
    text,
    string(rawEmail)).Scan(&id)
  if err != nil {
    log.Errorf("Messages SQL error: %v", err)
    return 0, err
  }
  if id == 0 {
    log.Errorf("Messages Not return last ID: %v", id)
    return 0, errors.New("Messages Not return last ID")
  }
  return id, nil
}