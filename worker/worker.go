package worker

import (
  "net/http"
  "strings"
  "fmt"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/parser"
  "github.com/le0pard/go-falcon/storage"
  "github.com/le0pard/go-falcon/spamassassin"
  "github.com/le0pard/go-falcon/clamav"
  "github.com/le0pard/go-falcon/redishook"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

// start worker
func startParserAndStorageWorker(config *config.Config, channel chan *smtpd.BasicEnvelope) {
  var (
    db          *storage.DBConn
    email       *parser.ParsedEmail
    report      string
    messageId   int
    err         error
  )
  settings := new(storage.AccountSettings)
  log.Debugf("Starting storage worker")
  for {
    envelop := <- channel
    // db connect
    db, err = storage.InitDatabase(config)
    if err != nil {
      log.Errorf("Couldn't connect to database: %v", err)
      continue
    } else {
      db.DB.SetMaxIdleConns(-1)
    }
    // get settings
    err = db.GetSettings(envelop.MailboxID, settings)
    if err != nil {
      // invalid settings
      continue
    }
    // parse email
    email, err = parser.ParseMail(envelop)
    if err == nil {
      messageId, err = db.StoreMail(email.MailboxID, email.Subject, email.Date, email.From.Address, email.From.Name, email.To.Address, email.To.Name, email.HtmlPart, email.TextPart, email.RawMail)
      // store attachments
      if err == nil {
        for _, attachment := range email.Attachments {
          _, err := db.StoreAttachment(email.MailboxID, messageId, attachment.AttachmentFileName, attachment.AttachmentType, attachment.AttachmentContentType, attachment.AttachmentContentID, attachment.AttachmentTransferEncoding, attachment.AttachmentBody)
          if err != nil {
            log.Errorf("StoreAttachment: %v", err)
          }
        }
      } else {
        log.Errorf("StoreMail: %v", err)
      }
      //cleanup messages
      db.CleanupMessages(email.MailboxID, settings.MaxMessages)
      // spamassassin
      if config.Spamassassin.Enabled {
        report, err = spamassassin.CheckSpamEmail(config, email.RawMail)
        if err == nil {
          // update spam info
          _, err = db.UpdateSpamReport(email.MailboxID, messageId, report)
          if err != nil {
            log.Errorf("UpdateSpamReport: %v", err)
          }
        } else {
          log.Errorf("CheckSpamEmail: %v", err)
        }
      }
      // clamav
      if config.Clamav.Enabled {
        report, err = clamav.CheckEmailForViruses(config, email.RawMail)
        if err == nil {
          if len(report) > 0 {
            // update viruses info
            _, err = db.UpdateVirusesReport(email.MailboxID, messageId, report)
            if err != nil {
              log.Errorf("UpdateVirusesReport: %v", err)
            }
          }
        } else {
          log.Errorf("CheckEmailForViruses: %v", err)
        }
      }
      // redis hooks
      if config.Redis.Enabled {
        redishook.SendNotifications(config, email.MailboxID, messageId, email.Subject)
      }
      // web hooks
      if config.Web_Hooks.Enabled {
        go webHookSender(config, email.MailboxID, messageId)
      }
    } else {
      log.Errorf("ParseMail: %v", err)
    }
    // cleanup
    db.Close() // force close the database, because pool is not under control
  }
}

// workers
func StartWorkers(config *config.Config, channel chan *smtpd.BasicEnvelope) {
  for i := 0; i < config.Storage.Pool; i++ {
    go startParserAndStorageWorker(config, channel)
  }
}

// web hooks

func webHookSender(config *config.Config, mailboxID, messageId int) {
  if len(config.Web_Hooks.Urls) > 0 {
    client := &http.Client{}
    for _, url := range config.Web_Hooks.Urls {
      r, err := http.NewRequest("POST", url,
        strings.NewReader(fmt.Sprintf("{\"channel\": \"/inboxes/%d\", \"ext\": {\"username\": \"%s\", \"password\": \"%s\"}, \"data\": {\"mailbox_id\": \"%d\", \"message_id\": \"%d\"}}", mailboxID, config.Web_Hooks.Username, config.Web_Hooks.Password, mailboxID, messageId)))

      if err != nil {
        log.Errorf("error init web hook: %v", err)
        continue
      } else {
        defer r.Body.Close()
        r.Header.Set("Content-Type", "application/json")
        resp, err := client.Do(r)
        if err != nil {
           log.Errorf("error init web hook: %v", err)
           continue
        }
        defer resp.Body.Close()
      }
    }
  }
}