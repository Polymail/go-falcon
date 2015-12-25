package worker

import (
	"github.com/le0pard/go-falcon/clamav"
	"github.com/le0pard/go-falcon/config"
	"github.com/le0pard/go-falcon/log"
	"github.com/le0pard/go-falcon/parser"
	"github.com/le0pard/go-falcon/protocol/smtpd"
	"github.com/le0pard/go-falcon/redisworker"
	"github.com/le0pard/go-falcon/spamassassin"
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
		inboxSettings, err := redisworker.GetCachedInboxSettings(config, envelop.MailboxID)
		if err != nil || 0 == inboxSettings.MaxMessages || 0 == inboxSettings.RateLimit {
			// inbox setting from database
			inboxSettings, err = config.DbPool.GeInboxSettings(envelop.MailboxID)
			// check settings
			if err != nil {
				// invalid settings
				continue
			} else {
				// cache setting in redis
				redisworker.StoreCachedInboxSettings(config, envelop.MailboxID, inboxSettings)
			}
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
				config.DbPool.CleanupMessages(email.MailboxID, inboxSettings)
				// redis counter
				if messageId > 0 && redisworker.IsNotSpamAttackCampaign(config, envelop.MailboxID) {
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
						redisworker.SendNotifications(config, email.MailboxID, messageId, email.Subject)
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
