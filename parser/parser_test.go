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
  {`Date: Sun, 04 Dec 2011 16:02:50 +0200
From: APP Error <sosedi@sosedi.ua>
To: app-support@sosedi.ua
Message-ID: <4edb7d8ae34d4_e7113fedc0834ecc846e@vnazarenko.mail>
Subject: [Sosedi2 production] cities#show (ActionView::Template::Error)
 "/Users/viktornazarenko/code/sosedi2/app/models/poll.rb:3...
Mime-Version: 1.0
Content-Type: text/plain;
 charset=UTF-8
Content-Transfer-Encoding: quoted-printable


=D0=A3=D0=BA=D0=B0=D0=B6=D0=B8=D1=82=D0=`,
  "[Sosedi2 production] cities#show (ActionView::Template::Error) \"/Users/viktornazarenko/code/sosedi2/app/models/poll.rb:3...", "app-support@sosedi.ua", "", "sosedi@sosedi.ua", "", ""},
  {`MIME-Version: 1.0
From: mainstay@sherwoodcompliance.co.uk
To: stephen.callaghan@greenfinch.ie
Date: 28 Jan 2013 16:27:28 +0000
Subject: test
Content-Type: multipart/mixed; boundary=--boundary_3_1c98cbdb-e45c-48ab-b94f-4a31cda787f6


----boundary_3_1c98cbdb-e45c-48ab-b94f-4a31cda787f6
Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: quoted-printable


----boundary_3_1c98cbdb-e45c-48ab-b94f-4a31cda787f6
Content-Type: unknown/unknown; name=OICLCostsPaymentProposal.csv
Content-Transfer-Encoding: base64
Content-Disposition: attachment

77u/U3VwcGxpZXJOYW1lfEJhdGNoUmVmZXJlbmNlfEluc3VyZXJSZWZlcmVuY2V8
R3Jvc3NTZXJ2aWNlUHJvdmlkZXJOZXR8U3VwcGxpZXJSZWZlcmVuY2V8VlJOfFBh
cnR5fEluY2lkZW50RGF0ZXxQb2xpY3lOdW1iZXJ8SW52b2ljZVR5cGV8VkFUTm90
QXBwbGljYWJsZXxSTVNMRmVlDQpHTEFTU3xzdGVwaGVuL2dsYXNzY2FyZTAwMXwz
Mi80MjY1fMKjMTY2LjgwfDk3MTc3NTF8VGV4dHx8MDEvMDEvMjAxMnx8R2xhc3Nj
YXJlfDU1LjYwfDE2LjgxDQoNCg==
----boundary_3_1c98cbdb-e45c-48ab-b94f-4a31cda787f6--`,
  "test", "stephen.callaghan@greenfinch.ie", "", "mainstay@sherwoodcompliance.co.uk", "", ""},
  {`Date: Sun, 31 Jul 2011 14:57:10 +0300
From: "Mr. Sender" <sender@mail.com>
To: Mr. X "wrongquote@b.com"
Message-ID: <4e35431682594_4ecd244557243018@hydra.mail>
Subject: illness
Mime-Version: 1.0
Content-Type: text/plain;
 charset=UTF-8
Content-Transfer-Encoding: 7bit

  illness 26 Dec - 26 Dec 2007`,
  "illness", "", "", "sender@mail.com", "", ""},
  {`Date: Sun, 31 Jul 2011 14:57:10 +0300
From: "Mr. Sender" <sender@mail.com>
To: aaaa@bbbbbb.com
Message-ID: <4e35431682594_4ecd244557243018@hydra.mail>
Subject: illness notification =?8bit?Q?ALPH=C3=89E?=
Mime-Version: 1.0
Content-Type: text/plain;
 charset=UTF-8
Content-Transfer-Encoding: 7bit

illness 26 Dec - 26 Dec 2007`,
  "illness notification =?8bit?Q?ALPH=C3=89E?=", "aaaa@bbbbbb.com", "", "sender@mail.com", "illness 26 Dec - 26 Dec 2007", ""},
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