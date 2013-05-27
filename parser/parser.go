package parser

import (
  "bytes"
  "net/mail"
  "mime"
  "mime/multipart"
  "io"
  "io/ioutil"
  "time"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

type ParsedEmail struct {
  env *smtpd.BasicEnvelope

  RawMail       []byte

  Subject       string
  Date          time.Time
  From          mail.Address
  To            mail.Address
  Headers       mail.Header

  EmailBody     []byte
}

// parse headers

func (email *ParsedEmail) parseEmailHeaders(msg *mail.Message) {
  var emailHeader = ""
  var err error

  email.Headers = msg.Header
  email.Subject = email.Headers.Get("Subject")
  email.Date, err = msg.Header.Date()
  if err != nil {
    log.Errorf("Failed parsing date: %v", err)
    email.Date = time.Now()
  }
  // from
  emailHeader = email.Headers.Get("From")
  if emailHeader != "" {
    fromEmail, err := mail.ParseAddress(emailHeader)
    if err != nil {
      email.From = mail.Address{ Address: email.env.From.Email() }
    } else {
      email.From = *fromEmail
    }
  }
  // to
  emailHeader = email.Headers.Get("To")
  if emailHeader != "" {
    toEmail, err := mail.ParseAddress(emailHeader)
    if err != nil {
      email.To = mail.Address{}
    } else {
      email.To = *toEmail
    }
  }
}

// parse plain email

func (email *ParsedEmail) parsePlainEmail() {
  contentType := email.Headers.Get("Content-Type")
  log.Debugf("ContentType: %v", contentType)
}

// parse plain email

func (email *ParsedEmail) parseMimeEmail() {
  contentType := email.Headers.Get("Content-Type")
  log.Debugf("ContentType: %v", contentType)
  contentTypeVal, contentTypeParams, err := mime.ParseMediaType(contentType)
  if err != nil {
    log.Errorf("Invalid ContentType: %v", err)
    return
  }
  log.Debugf("ContentType Value: %v", contentTypeVal)
  log.Debugf("ContentType params: %v", contentTypeParams)
  if contentTypeParams["boundary"] == "" {
    log.Errorf("No boundary: %v", contentTypeParams)
    return
  }
  buf := new(bytes.Buffer)
  bodyReader := bytes.NewReader(email.EmailBody)
  reader := multipart.NewReader(bodyReader, contentTypeParams["boundary"])

  for {
    p, err := reader.NextPart()
    if err == io.EOF {
      break
    }
    if err != nil {
      log.Errorf("NextPart: %v", err)
    }
    pbody, err := ioutil.ReadAll(p)

  }

  part, err := reader.NextPart()
  if part == nil || err != nil {
    log.Errorf("No boundary: %v", contentTypeParams)
    return
  }
  if _, err := io.Copy(buf, part); err != nil {
    log.Errorf("Error: %v", err)
  }
  log.Debugf("Part: %v", buf.String())
}

// parse body

func (email *ParsedEmail) parseEmailBody(msg *mail.Message, body []byte) {
  email.EmailBody = body
  mimeVersion := email.Headers.Get("Mime-Version")
  if mimeVersion != "" {
    email.parseMimeEmail()
  } else {
    email.parsePlainEmail()
  }
}

// obj

type EmailParser struct {
}

// parse email

func (parser *EmailParser) ParseMail(env *smtpd.BasicEnvelope) {
  email := ParsedEmail{ env: env }
  msg, err := mail.ReadMessage(bytes.NewBuffer(email.env.MailBody))
  if err != nil {
    log.Errorf("Failed parsing message: %v", err)
    return
  }
  mailBody, err := ioutil.ReadAll(msg.Body)
  if err != nil {
    log.Errorf("Failed parsing message: %v", err)
    return
  }
  email.RawMail = email.env.MailBody
  email.parseEmailHeaders(msg)
  email.parseEmailBody(msg, mailBody)
  //log.Debugf("parsed: %v", email)
}