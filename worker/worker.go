package worker

import (
  "fmt"
  "github.com/le0pard/go-falcon/clamav"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/parser"
  "github.com/le0pard/go-falcon/protocol/smtpd"
  "github.com/le0pard/go-falcon/redishook"
  "github.com/le0pard/go-falcon/spamassassin"
  "net/http"
  "strings"
)

// start worker
func startParserAndStorageWorker(config *config.Config, channel chan *smtpd.BasicEnvelope) {
  var (
    email     *parser.ParsedEmail
    report    string
    messageId int
  )
  log.Debugf("Starting storage worker")
  for {
    envelop := <-channel
    // get settings
    maxMessages, err := config.DbPool.GetMaxMessages(envelop.MailboxID)
    if err != nil {
      // invalid settings
      continue
    }
    // parse email
    email, err = parser.ParseMail(envelop)
    if err == nil {
      messageId, err = config.DbPool.StoreMail(email.MailboxID, email.Subject, email.Date, email.From.Address, email.From.Name, email.To.Address, email.To.Name, email.HtmlPart, email.TextPart, email.RawMail)
      // store attachments
      if err == nil {
        for _, attachment := range email.Attachments {
          _, err := config.DbPool.StoreAttachment(email.MailboxID, messageId, attachment.AttachmentFileName, attachment.AttachmentType, attachment.AttachmentContentType, attachment.AttachmentContentID, attachment.AttachmentTransferEncoding, attachment.AttachmentBody)
          if err != nil {
            log.Errorf("StoreAttachment: %v", err)
          }
        }

        //cleanup messages
        config.DbPool.CleanupMessages(email.MailboxID, maxMessages)
        // redis counter
        if messageId > 0 && redishook.IsNotSpamAttackCampaign(config, envelop.MailboxID) {
          // spamassassin
          if config.Spamassassin.Enabled {
            report, err = spamassassin.CheckSpamEmail(config, email.RawMail)
            if err == nil {
              // update spam info
              _, err = config.DbPool.UpdateSpamReport(email.MailboxID, messageId, report)
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
                _, err = config.DbPool.UpdateVirusesReport(email.MailboxID, messageId, report)
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
        }

      } else {
        log.Errorf("StoreMail: %v", err)
      }

    } else {
      log.Errorf("ParseMail: %v", err)
    }
  }
}

// workers
func StartWorkers(config *config.Config, channel chan *smtpd.BasicEnvelope) {
  for i := 0; i < config.Adapter.Workers_Size; i++ {
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
