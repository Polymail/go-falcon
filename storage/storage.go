package storage

import (
  "strings"
  "errors"
  "database/sql"
  "strconv"
  "time"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/utils"
  "github.com/le0pard/go-falcon/storage/postgresql"
)

type AccountSettings struct {
  MaxMessages int
}

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
  err := db.DB.QueryRow(db.config.Storage.Auth_Sql, username).Scan(&id, &password)
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

// save email

func (db *DBConn) StoreMail(mailboxId int, subject string, date time.Time, from, from_name, to, to_name, html, text string, rawEmail []byte) (int, error) {
  var (
    id int
  )
  sql := strings.Replace(db.config.Storage.Messages_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
  err := db.DB.QueryRow(sql,
    mailboxId,
    subject,
    date,
    from,
    from_name,
    to,
    to_name,
    html,
    text,
    string(rawEmail),
    len(rawEmail)).Scan(&id)
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

// update spam report

func(db *DBConn) UpdateSpamReport(mailboxId int, messageId int, spamReport string) (int, error) {
  var (
    id int
  )
  sql := strings.Replace(db.config.Storage.Spamassassin_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
  err := db.DB.QueryRow(sql,
    mailboxId,
    messageId,
    spamReport).Scan(&id)
  if err != nil {
    log.Errorf("Spamassassin SQL error: %v", err)
    return 0, err
  }
  return id, nil
}


// save attachment

func (db *DBConn) StoreAttachment(mailboxId int, messageId int, filename, attachmentType, contentType, contentId, transferEncoding string, rawData []byte) (int, error) {
  var (
    id int
  )
  sql := strings.Replace(db.config.Storage.Attachments_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
  err := db.DB.QueryRow(sql,
    mailboxId,
    messageId,
    filename,
    attachmentType,
    contentType,
    contentId,
    transferEncoding,
    rawData,
    len(rawData)).Scan(&id)
  if err != nil {
    log.Errorf("Attachments SQL error: %v", err)
    return 0, err
  }
  if id == 0 {
    log.Errorf("Attachments Not return last ID: %v", id)
    return 0, errors.New("Attachments Not return last ID")
  }
  return id, nil
}

// get settings

func (db *DBConn) GetSettings(mailboxId int) (*AccountSettings, error) {
  settings := &AccountSettings{}
  err := db.DB.QueryRow(db.config.Storage.Settings_Sql, mailboxId).Scan(&settings.MaxMessages)
  if err != nil {
    log.Errorf("Settings SQL error: %v", err)
    return nil, err
  }
  return settings, err
}

// cleanup messages
func (db *DBConn) CleanupMessages(mailboxId, maxMessages int) error {
  if db.config.Storage.Max_Messages_Enabled && maxMessages > 0 {
    var (
      count int
      tmpId int
      msgIds []string
    )
    sql := strings.Replace(db.config.Storage.Max_Messages_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
    err := db.DB.QueryRow(sql, mailboxId).Scan(&count)
    if err != nil {
      log.Errorf("CleanupMessages SQL error: %v", err)
      return err
    }
    cleanupCount := count - maxMessages
    if cleanupCount > 0 {
      sql := strings.Replace(db.config.Storage.Max_Messages_Cleanup_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
      rows, err := db.DB.Query(sql, mailboxId, cleanupCount)
      if err != nil {
        log.Errorf("CleanupMessages SQL error: %v", err)
        return err
      }
      defer rows.Close()
      for rows.Next() {
        err := rows.Scan(&tmpId)
        if err != nil {
          log.Errorf("CleanupMessages SQL error: %v", err)
          return err
        }
        msgIds = append(msgIds, strconv.Itoa(tmpId))
      }
      if len(msgIds) > 0 {
        sql := strings.Replace(db.config.Storage.Max_Attachments_Cleanup_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
        for _, msgId := range msgIds {
          _, err := db.DB.Query(sql, mailboxId, msgId)
          if err != nil {
            log.Errorf("CleanupMessages SQL error: %v", err)
            return err
          }
        }
      }
    }
  }
  return nil
}

// close connection

func (db *DBConn) Close() {
  db.DB.Close()
}
