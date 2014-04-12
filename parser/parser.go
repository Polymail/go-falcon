package parser

import (
  "bytes"
  "net/mail"
  "net/textproto"
  "mime"
  "mime/multipart"
  "io"
  "io/ioutil"
  "strings"
  "time"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

type ParsedAttachment struct {
  AttachmentType                string
  AttachmentFileName            string
  AttachmentTransferEncoding    string
  AttachmentContentType         string
  AttachmentContentID           string
  AttachmentBody                string
}

type ParsedEmail struct {
  env *smtpd.BasicEnvelope
  MailboxID     int
  RawMail       []byte

  Subject       string
  Date          time.Time
  From          mail.Address
  To            mail.Address
  Headers       mail.Header

  HtmlPart      string
  TextPart      string

  Attachments   []ParsedAttachment

  EmailBody     []byte
}

// parse headers

func (email *ParsedEmail) parseEmailHeaders(msg *mail.Message) {
  var emailHeader = ""
  var err error

  email.Headers = msg.Header
  email.Subject = MimeHeaderDecode(email.Headers.Get("Subject"))
  email.Date, err = msg.Header.Date()
  if err != nil {
    email.Date = time.Now()
  }
  // from
  emailHeader = email.Headers.Get("From")
  if emailHeader != "" {
    fromEmail, err := mail.ParseAddress(emailHeader)
    if err != nil {
      if email.env.From.Email() != "" {
        email.From = mail.Address{ Address: email.env.From.Email() }
      } else {
        email.From = mail.Address{ Address: getInvalidFromToHeader(emailHeader) }
      }
    } else {
      email.From = *fromEmail
    }
  }
  // to
  emailHeader = email.Headers.Get("To")
  if emailHeader != "" {
    toEmail, err := mail.ParseAddress(emailHeader)
    if err != nil {
      if (len(email.env.Rcpts) > 0){
        email.To = mail.Address{ Address: email.env.Rcpts[0].Email() }
      } else {
        email.To = mail.Address{ Address: getInvalidFromToHeader(emailHeader) }
      }
    } else {
      email.To = *toEmail
    }
  }
}

// select type of email

func (email *ParsedEmail) parseEmailByType(headers textproto.MIMEHeader, pbody []byte) {
  var (
    contentDispositionVal string
    contentDispositionParams map[string]string
  )

  contentType, contentDisposition, contentTransferEncoding := headers.Get("Content-Type"), headers.Get("Content-Disposition"), headers.Get("Content-Transfer-Encoding")
  if contentType == "" {
    contentType = "text/plain; charset=UTF-8"
  }
  if contentTransferEncoding == "" {
    contentTransferEncoding = "8bit"
  }

  // content type
  contentType = FixMailEncodedHeader(contentType)
  contentTypeVal, contentTypeParams, err := mime.ParseMediaType(contentType)
  if err != nil {
    log.Errorf("Invalid ContentType: %v", err)
    return
  }
  contentTypeVal = strings.ToLower(contentTypeVal)

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
  case "text/plain":
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
  case "message/delivery-status", "text/rfc822-headers":
    email.TextPart = email.TextPart + FixEncodingAndCharsetOfPart(string(pbody), contentTransferEncoding, contentTypeParams["charset"], true)
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
      attachment := ParsedAttachment{ AttachmentType: contentDispositionVal, AttachmentFileName: filename, AttachmentBody: FixEncodingAndCharsetOfPart(string(pbody), contentTransferEncoding, contentTypeParams["charset"], false), AttachmentContentType: contentTypeVal, AttachmentTransferEncoding: contentTransferEncoding, AttachmentContentID: attachmentContentID }
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

func (email *ParsedEmail) parseEmailPart(part *multipart.Part) {
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
  reader := multipart.NewReader(bodyReader, boundary)

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
  email := &ParsedEmail{ env: env, MailboxID: env.MailboxID }
  msg, err := mail.ReadMessage(bytes.NewBuffer(email.env.MailBody))
  if err != nil {
    log.Errorf("Failed parsing message: %v", err)
    return nil, err
  }
  mailBody, err := ioutil.ReadAll(msg.Body)
  if err != nil {
    log.Errorf("Failed parsing message: %v", err)
    return nil, err
  }
  email.RawMail = email.env.MailBody
  email.parseEmailHeaders(msg)
  email.parseEmailBody(mailBody)
  return email, nil
}
