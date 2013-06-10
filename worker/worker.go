package worker

import (
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/parser"
  "github.com/le0pard/go-falcon/storage"
  "github.com/le0pard/go-falcon/spamassassin"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

// start worker
func startParserAndStorageWorker(config *config.Config, channel chan *smtpd.BasicEnvelope) {
  log.Debugf("Starting storage worker")
  emailParser := parser.EmailParser{}
  db, err := storage.InitDatabase(config)
  if err != nil {
    log.Errorf("Couldn't connect to database: %v", err)
    return
  }
  for {
    envelop := <- channel
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
          db.StoreAttachment(email.MailboxID, messageId, attachment.AttachmentFileName, attachment.AttachmentType, attachment.AttachmentContentType, attachment.AttachmentContentID, attachment.AttachmentTransferEncoding, attachment.AttachmentBody)
        }
      }
      //cleanup messages
      db.CleanupMessages(email.MailboxID, settings.MaxMessages)
      // spamassassin
      if config.Spamassassin.Enabled {
        spamassassin.CheckSpamEmail(config, email.RawMail)
      }
    }
  }
}

// workers
func StartWorkers(config *config.Config, channel chan *smtpd.BasicEnvelope) {
  for i := 0; i < config.Storage.Pool; i++ {
    go startParserAndStorageWorker(config, channel)
  }
}
