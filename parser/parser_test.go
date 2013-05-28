package parser

import (
  "github.com/le0pard/go-falcon/protocol/smtpd"
  "testing"
)

// good mails

type goodMailTypeTest struct {
  RawBody     string
}

var goodMailTypeTests = []goodMailTypeTest{
  {`
From: Private Person <me@fromdomain.com>
To: A Test User <test@todomain.com>
CC: <test2@todomain.com>
CC: <test3@todomain.com>
Subject: SMTP e-mail test

This is a test e-mail message.
  `},
}

// bad mails

type badMailTypeTest struct {
  RawBody     string
}


type stubEnvelop struct {
  env struct {
    MailboxID    int
    MailBody     []byte
  }
}


func testMailParser(t *testing.T) {
  emailParser := EmailParser{}
  for _, mail := range goodMailTypeTests {
    envelop := &smtpd.BasicEnvelope{ MailboxID: 0, MailBody: []byte(mail.RawBody)} 
    emailParser.ParseMail(envelop)
  }
}

