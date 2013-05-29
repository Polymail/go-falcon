package parser

import (
  "github.com/le0pard/go-falcon/protocol/smtpd"
  "encoding/json"
  "strings"
  "testing"
)

// good mails

type goodMailTypeTest struct {
  RawBody     string

  Subject     string
  To          string
  ToName      string
  From        string
  Text        string
  Html        string
}

var goodMailTypeTests = []goodMailTypeTest{
  {`From: Private Person <me@fromdomain.com>
To: A Test User <test@todomain.com>
CC: <test2@todomain.com>
CC: <test3@todomain.com>
Subject: SMTP e-mail test

This is a test e-mail message.`,
  "SMTP e-mail test", "test@todomain.com", "A Test User", "me@fromdomain.com", "This is a test e-mail message.", ""},
}

// bad mails

type badMailTypeTest struct {
  RawBody     string
}


func escapeString(v string) string {
  bytes, _ := json.Marshal(v)
  return string(bytes)
}

func expectEq(t *testing.T, expected, actual, what string) {
  if expected == actual {
    return
  }
  t.Errorf("Unexpected value for %s; got %s (len %d) but expected: %s (len %d)",
    what, escapeString(actual), len(actual), escapeString(expected), len(expected))
}


func TestMailParser(t *testing.T) {
  emailParser := EmailParser{}
  for _, mail := range goodMailTypeTests {
    testBody := strings.Replace(mail.RawBody, "\n", "\r\n", -1)
    // parse email
    envelop := &smtpd.BasicEnvelope{ MailboxID: 0, MailBody: []byte(testBody)}
    email, err := emailParser.ParseMail(envelop)
    if email == nil || err != nil {
      t.Error("Error in parsing email: %v", err)
    } else {
      expectEq(t, mail.Subject, email.Subject, "Value of subject")
      expectEq(t, mail.To, email.To.Address, "Value of to email")
      expectEq(t, mail.ToName, email.To.Name, "Value of to email name")
      expectEq(t, mail.From, email.From.Address, "Value of from email")
      expectEq(t, mail.Text, email.TextPart, "Value of text")
      expectEq(t, mail.Html, email.HtmlPart, "Value of html")
    }
  }
}