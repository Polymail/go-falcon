package worker

import (
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/parser"
  "github.com/le0pard/go-falcon/storage"
  "github.com/le0pard/go-falcon/spamassassin"
  "github.com/le0pard/go-falcon/clamav"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

// start worker
func startParserAndStorageWorker(config *config.Config, channel chan *smtpd.BasicEnvelope) {
  log.Debugf("Starting storage worker")
  emailParser := parser.EmailParser{}
  for {
    envelop := <- channel
    // db connect 
    db, err := storage.InitDatabase(config)
    if err != nil {
      log.Errorf("Couldn't connect to database: %v", err)
      continue
    } else {
      db.DB.SetMaxIdleConns(-1)
    }
    // get settings
    settings, err := db.GetSettings(envelop.MailboxID)
    if err != nil {
      // invalid settings
      continue
    }
    // parse email
    email, err := emailParser.ParseMail(envelop)
    if err == nil {
      messageId, err := db.StoreMail(email.MailboxID, email.Subject, email.Date, email.From.Address, email.From.Name, email.To.Address, email.To.Name, email.HtmlPart, email.TextPart, email.RawMail)
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
        spamassassinJson, err := spamassassin.CheckSpamEmail(config, email.RawMail)
        if err == nil {
          // update spam info
          _, err := db.UpdateSpamReport(email.MailboxID, messageId, spamassassinJson)
          if err != nil {
            log.Errorf("UpdateSpamReport: %v", err)
          }
        } else {
          log.Errorf("CheckSpamEmail: %v", err)
        }
      }
      // clamav
      if config.Clamav.Enabled {
        clamavReport, err := clamav.CheckEmailForViruses(config, email.RawMail)
        if err == nil {
          log.Debugf("Spam: %s", clamavReport)
          // update viruses info
          //_, err := db.UpdateVirusesReport(email.MailboxID, messageId, spamassassinJson)
          //if err != nil {
          //  log.Errorf("UpdateVirusesReport: %v", err)
          //}
        } else {
          log.Errorf("CheckEmailForViruses: %v", err)
        }
      }
    } else {
      log.Errorf("ParseMail: %v", err)
    }
    db.Close()
  }
}

// workers
func StartWorkers(config *config.Config, channel chan *smtpd.BasicEnvelope) {
  for i := 0; i < config.Storage.Pool; i++ {
    go startParserAndStorageWorker(config, channel)
  }
}
