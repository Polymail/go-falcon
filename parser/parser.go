package parser

import (
	"bytes"
	"github.com/Polymail/go-falcon/go_multipart_pacthed"
	"github.com/Polymail/go-falcon/log"
	"github.com/Polymail/go-falcon/protocol/smtpd"
	"io"
	"io/ioutil"
	"mime"
	"net/mail"
	"net/textproto"
	"strings"
	"time"
)

type ParsedAttachment struct {
	AttachmentType             string
	AttachmentFileName         string
	AttachmentTransferEncoding string
	AttachmentContentType      string
	AttachmentContentID        string
	AttachmentBody             string
}

type ParsedEmail struct {
	env       *smtpd.BasicEnvelope
	MailboxID int
	RawMail   []byte

	Subject string
	Date    time.Time
	From    mail.Address
	To      mail.Address
	Headers mail.Header

	HtmlPart string
	TextPart string

	Attachments []ParsedAttachment

	EmailBody []byte
}

// parse headers

func extractFromToHeader(header string) mail.Address {
	name, address := getInvalidFromToHeader(header)
	return mail.Address{Name: name, Address: address}
}

func getFromOrToHeader(email *ParsedEmail, headerType string) mail.Address {
	mailAddressRes := mail.Address{}

	emailHeader := email.Headers.Get(headerType)
	if emailHeader != "" {
		toEmail, err := mail.ParseAddress(emailHeader)
		if err != nil {
			mailAddressRes = extractFromToHeader(emailHeader)
		} else {
			mailAddressRes = *toEmail
			mailAddressRes.Name = MimeHeaderDecode(mailAddressRes.Name)
		}
	}

	return mailAddressRes
}

func (email *ParsedEmail) parseEmailHeaders(msg *mail.Message) {
	var err error

	email.Headers = msg.Header
	email.Subject = MimeHeaderDecode(email.Headers.Get("Subject"))
	email.Date, err = msg.Header.Date()
	if err != nil {
		email.Date = time.Now()
	}
	if email.Date.Year() < 1970 {
		email.Date = time.Now()
	}
	// from
	email.From = getFromOrToHeader(email, "From")
	// to
	email.To = getFromOrToHeader(email, "To")
}

// select type of email

func (email *ParsedEmail) parseEmailByType(headers textproto.MIMEHeader, pbody []byte) {
	var (
		contentDispositionVal    string
		contentDispositionParams map[string]string
	)

	contentType, contentDisposition, contentTransferEncoding := headers.Get("Content-Type"), headers.Get("Content-Disposition"), headers.Get("Content-Transfer-Encoding")
	if contentType == "" {
		contentType = "text/plain; charset=UTF-8"
	}
	if contentTransferEncoding == "" {
		contentTransferEncoding = "8bit"
	} else {
		contentTransferEncoding = strings.ToLower(contentTransferEncoding)
	}

	// content type
	contentType = FixMailEncodedHeader(contentType)
	contentTypeVal, contentTypeParams, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Errorf("Invalid ContentType: %v", err)
		return
	} else {
		contentTypeVal = strings.ToLower(contentTypeVal)
	}

	// content disposition
	if contentDisposition != "" {
		contentDisposition = FixMailEncodedHeader(contentDisposition)
		contentDispositionVal, contentDispositionParams, err = mime.ParseMediaType(contentDisposition)
		if err != nil {
			log.Errorf("Invalid ContentDisposition: %v", err)
			return
		}
		contentDispositionVal = strings.ToLower(contentDispositionVal)
		if contentDispositionVal == "attachment" {
			email.parseAttachment(headers, contentTypeVal, contentDispositionVal, contentTransferEncoding, contentTypeParams, contentDispositionParams, pbody)
			return
		}
	}

	// contentType cases
	switch contentTypeVal {
	case "text/html":
		email.HtmlPart = email.HtmlPart + FixEncodingAndCharsetOfPart(string(pbody), contentTransferEncoding, contentTypeParams["charset"], true)
	case "text/plain", "message/delivery-status", "text/rfc822-headers":
		email.TextPart = email.TextPart + FixEncodingAndCharsetOfPart(string(pbody), contentTransferEncoding, contentTypeParams["charset"], true)
	case "message/rfc822":
		msg, err := mail.ReadMessage(bytes.NewBuffer(pbody))
		if err != nil {
			log.Errorf("Failed parsing message of rfc822: %v", err)
		} else {
			mailBody, err := ioutil.ReadAll(msg.Body)
			if err != nil {
				log.Errorf("Failed parsing message of rfc822: %v", err)
			} else {
				email.Headers = msg.Header
				email.parseEmailBody(mailBody)
			}
		}
	default:
		// multipart
		if strings.HasPrefix(contentTypeVal, "multipart/") {
			email.parseMimeEmail(pbody, contentTypeParams["boundary"])
		} else if contentDisposition != "" {
			email.parseAttachment(headers, contentTypeVal, contentDispositionVal, contentTransferEncoding, contentTypeParams, contentDispositionParams, pbody)
			// attachments without content disposition (sic!)
		} else if strings.HasPrefix(contentTypeVal, "image/") || strings.HasPrefix(contentTypeVal, "audio/") || strings.HasPrefix(contentTypeVal, "video/") || strings.HasPrefix(contentTypeVal, "application/") || strings.HasPrefix(contentTypeVal, "text/") {
			email.parseAttachment(headers, contentTypeVal, "attachment", contentTransferEncoding, contentTypeParams, contentDispositionParams, pbody)
		} else {
			log.Errorf("Unknown content type: %s", contentTypeVal)
			log.Errorf("Unknown content params: %v", contentTypeParams)
			log.Errorf("Unknown content: %v", string(pbody))
		}
	}
}

// parse attachments

func (email *ParsedEmail) parseAttachment(headers textproto.MIMEHeader, contentTypeVal, contentDispositionVal, contentTransferEncoding string, contentTypeParams, contentDispositionParams map[string]string, pbody []byte) {
	switch contentDispositionVal {
	case "attachment", "inline":
		filename := getFilenameOfAttachment(contentTypeParams, contentDispositionParams)
		attachmentContentID := headers.Get("Content-ID")
		if attachmentContentID != "" {
			contentId, err := mail.ParseAddress(attachmentContentID)
			if err == nil {
				attachmentContentID = contentId.Address
			} else {
				attachmentContentID = getInvalidContentId(attachmentContentID)
			}
		}
		attachment := ParsedAttachment{AttachmentType: contentDispositionVal, AttachmentFileName: filename, AttachmentBody: FixEncodingAndCharsetOfPart(string(pbody), contentTransferEncoding, contentTypeParams["charset"], false), AttachmentContentType: contentTypeVal, AttachmentTransferEncoding: contentTransferEncoding, AttachmentContentID: attachmentContentID}
		email.Attachments = append(email.Attachments, attachment)
	default:
		log.Errorf("Unknown content disposition: %s", contentDispositionVal)
		log.Errorf("Unknown content params: %v", contentDispositionParams)
	}
}

// get filename of attachment

func getFilenameOfAttachment(contentTypeParams, contentDispositionParams map[string]string) string {
	filename := ""
	if contentTypeParams != nil {
		if filename == "" {
			filename = contentTypeParams["filename"]
		}
		if filename == "" {
			filename = contentTypeParams["name"]
		}
	}
	if contentDispositionParams != nil && filename == "" {
		if filename == "" {
			filename = contentDispositionParams["filename"]
		}
		if filename == "" {
			filename = contentDispositionParams["name"]
		}
	}
	filename = MimeHeaderDecode(filename)
	return filename
}

// parse part of email

func (email *ParsedEmail) parseEmailPart(part *go_multipart_pacthed.Part) {
	pbody, err := ioutil.ReadAll(part)
	if err != nil {
		log.Errorf("Read part: %v", err)
		return
	}
	email.parseEmailByType(part.Header, pbody)
}

// parse plain email

func (email *ParsedEmail) parsePlainEmail() {
	email.parseEmailByType(textproto.MIMEHeader(email.Headers), email.EmailBody)
}

// parse plain email

func (email *ParsedEmail) parseMimeEmail(pbody []byte, boundary string) {
	if boundary == "" {
		log.Errorf("Doesn't found boundary in MIME: %s", boundary)
		return
	}

	bodyReader := bytes.NewReader(pbody)
	reader := go_multipart_pacthed.NewReader(bodyReader, boundary)

	for {
		p, err := reader.NextPart()

		if err != nil {
			if io.EOF != err {
				log.Errorf("Mime Part error: %v", err)
			}
			break
		} else {
			email.parseEmailPart(p)
		}
	}
}

// parse body

func (email *ParsedEmail) parseEmailBody(body []byte) {
	email.EmailBody = body
	mimeVersion := email.Headers.Get("Mime-Version")
	contentType := email.Headers.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain; charset=UTF-8"
	}
	contentTypeVal, contentTypeParams, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Errorf("Invalid ContentType: %v", err)
		return
	}
	if mimeVersion != "" && strings.HasPrefix(strings.ToLower(contentTypeVal), "multipart/") && contentTypeParams["boundary"] != "" {
		email.parseMimeEmail(email.EmailBody, contentTypeParams["boundary"])
	} else {
		email.parsePlainEmail()
	}
}

// parse email

func ParseMail(env *smtpd.BasicEnvelope) (*ParsedEmail, error) {
	email := &ParsedEmail{env: env, MailboxID: env.MailboxID}
	msg, err := mail.ReadMessage(bytes.NewBuffer(email.env.MailBody))
	if err != nil {
		log.Errorf("Failed parsing ReadMessage: %v", err)
		return nil, err
	}
	mailBody, err := ioutil.ReadAll(msg.Body)
	if err != nil {
		log.Errorf("Failed parsing ReadAll: %v", err)
		return nil, err
	}
	email.RawMail = email.env.MailBody
	email.parseEmailHeaders(msg)
	email.parseEmailBody(mailBody)
	return email, nil
}
