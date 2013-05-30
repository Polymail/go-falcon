package parser

import (
  "bytes"
  "net/mail"
  "net/textproto"
  "mime"
  "mime/multipart"
  "encoding/base64"
  "regexp"
  "io"
  "io/ioutil"
  "strings"
  "time"
  "github.com/qiniu/iconv"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

type ParsedAttachment struct {
  AttachmentType                string
  AttachmentFileName            string
  AttachmentTransferEncoding    string
  AttachmentContentType         string
  AttachmentBody                []byte
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
  email.Subject = email.Headers.Get("Subject")
  email.Date, err = msg.Header.Date()
  if err != nil {
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
      if (len(email.env.Rcpts) > 0){
        email.To = mail.Address{ Address: email.env.Rcpts[0].Email() }
      } else {
        email.To = mail.Address{}
      }
    } else {
      email.To = *toEmail
    }
  }
}

// select type of email

func (email *ParsedEmail) parseEmailByType(headers textproto.MIMEHeader, pbody []byte) {
  contentType, contentDisposition, contentTransferEncoding := headers.Get("Content-Type"), headers.Get("Content-Disposition"), headers.Get("Content-Transfer-Encoding")
  if contentType == "" {
    contentType = "text/plain; charset=UTF-8"
  }
  if contentTransferEncoding == "" {
    contentTransferEncoding = "8bit"
  }
  contentTypeVal, contentTypeParams, err := mime.ParseMediaType(contentType)
  if err != nil {
    log.Errorf("Invalid ContentType: %v", err)
    return
  }
  if contentDisposition != "" {
    contentDispositionVal, contentDispositionParams, err := mime.ParseMediaType(contentDisposition)
    if err != nil {
      log.Errorf("Invalid ContentDisposition: %v", err)
      return
    }
    switch strings.ToLower(contentDispositionVal) {
    case "attachment", "inline":
      filename := contentDispositionParams["filename"]
      if filename == "" {
        filename = contentDispositionParams["name"]
      }
      if filename == "" {
        filename = contentTypeParams["filename"]
      }
      attachmentByte := []byte(mailTransportDecode(string(pbody), contentTransferEncoding, contentTypeParams["charset"]))
      attachment := ParsedAttachment{ AttachmentType: contentDispositionVal, AttachmentFileName: filename, AttachmentBody: attachmentByte, AttachmentContentType: contentTypeVal, AttachmentTransferEncoding: contentTransferEncoding }
      email.Attachments = append(email.Attachments, attachment)
    default:
      log.Errorf("Unknown content disposition: %s", contentDispositionVal)
      log.Errorf("Unknown content params: %v", contentDispositionParams)
    }
  } else {
    switch strings.ToLower(contentTypeVal) {
    case "text/html":
      email.HtmlPart = mailTransportDecode(string(pbody), contentTransferEncoding, contentTypeParams["charset"])
    case "text/plain":
      email.TextPart = mailTransportDecode(string(pbody), contentTransferEncoding, contentTypeParams["charset"])
    default:
      if strings.HasPrefix(strings.ToLower(contentTypeVal), "multipart/") {
        email.parseMimeEmail(pbody, contentTypeParams["boundary"])
      } else {
        log.Errorf("Unknown content type: %s", contentTypeVal)
        log.Errorf("Unknown content params: %v", contentTypeParams)
        log.Errorf("Unknown content: %v", string(pbody))
      }
    }
  }
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
    if err == io.EOF {
      break
    }
    if err != nil {
      log.Errorf("Mime Part error: %v", err)
    } else {
      email.parseEmailPart(p)
    }
  }
}

// parse body

func (email *ParsedEmail) parseEmailBody(msg *mail.Message, body []byte) {
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

// decode from 7bit to 8bit UTF-8
func mailTransportDecode(pbody string, encodingType string, charset string) string {
  if charset == "" {
    charset = "UTF-8"
  } else {
    charset = strings.ToUpper(charset)
  }
  if strings.ToLower(encodingType) == "base64" {
    pbody = decodeFromBase64(pbody)
  }
  if charset != "UTF-8" {
    charset = decodeFixCharset(charset)
    cd, err := iconv.Open(charset, "UTF-8")
    if err != nil {
      return pbody
    }
    defer cd.Close()
    return cd.ConvString(pbody)
  }
  return pbody
}

func decodeFromBase64(data string) string {
  buf := bytes.NewBufferString(data)
  decoder := base64.NewDecoder(base64.StdEncoding, buf)
  res, _ := ioutil.ReadAll(decoder)
  return string(res)
}

func decodeFixCharset(charset string) string {
  reg, _ := regexp.Compile(`[_:.\/\\]`)
  fixedCharset := reg.ReplaceAllString(charset, "-")
  // Fix charset
  // borrowed from http://squirrelmail.svn.sourceforge.net/viewvc/squirrelmail/trunk/squirrelmail/include/languages.php?revision=13765&view=markup
  // OE ks_c_5601_1987 > cp949
  fixedCharset = strings.Replace(fixedCharset, "ks-c-5601-1987", "cp949", -1)
  // Moz x-euc-tw > euc-tw
  fixedCharset = strings.Replace(fixedCharset, "x-euc", "euc", -1)
  // Moz x-windows-949 > cp949
  fixedCharset = strings.Replace(fixedCharset, "x-windows_", "cp", -1)
  // windows-125x and cp125x charsets
  fixedCharset = strings.Replace(fixedCharset, "windows-", "cp", -1)
  // ibm > cp
  fixedCharset = strings.Replace(fixedCharset, "ibm", "cp", -1)
  // iso-8859-8-i -> iso-8859-8
  fixedCharset = strings.Replace(fixedCharset, "iso-8859-8-i", "iso-8859-8", -1)
  if charset != fixedCharset {
    return fixedCharset
  }
  return charset
}

// obj

type EmailParser struct {
}

// parse email

func (parser *EmailParser) ParseMail(env *smtpd.BasicEnvelope) (*ParsedEmail, error) {
  email := ParsedEmail{ env: env, MailboxID: env.MailboxID  }
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
  email.parseEmailBody(msg, mailBody)
  return &email, nil
}
