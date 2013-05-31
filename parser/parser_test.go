package parser

import (
  stdlog "log"
  "os"
  "encoding/json"
  "strings"
  "testing"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

// good mails

type goodMailAttachments struct {
  Filename      string
}

type goodMailTypeTest struct {
  RawBody     string

  Subject     string
  To          string
  ToName      string
  From        string
  FromName    string
  Text        string
  Html        string

  Attachments  []goodMailAttachments
}

var goodMailTypeTests = []goodMailTypeTest{
  {`From: Private Person <me@fromdomain.com>
To: A Test User <test@todomain.com>
CC: <test2@todomain.com>
CC: <test3@todomain.com>
Subject: SMTP e-mail test

This is a test e-mail message.`,
  "SMTP e-mail test", "test@todomain.com", "A Test User", "me@fromdomain.com", "Private Person", "This is a test e-mail message.", "", []goodMailAttachments{}},

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
  "[Sosedi2 production] cities#show (ActionView::Template::Error) \"/Users/viktornazarenko/code/sosedi2/app/models/poll.rb:3...", "app-support@sosedi.ua", "", "sosedi@sosedi.ua", "APP Error", "\n=D0=A3=D0=BA=D0=B0=D0=B6=D0=B8=D1=82=D0=", "", []goodMailAttachments{}},

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
  "test", "stephen.callaghan@greenfinch.ie", "", "mainstay@sherwoodcompliance.co.uk", "", "", "", []goodMailAttachments{
    {"OICLCostsPaymentProposal.csv"},
  }},

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
  "illness", "", "", "sender@mail.com", "", "  illness 26 Dec - 26 Dec 2007", "", []goodMailAttachments{}},

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
  "illness notification ALPHÉE", "aaaa@bbbbbb.com", "", "sender@mail.com", "Mr. Sender", "illness 26 Dec - 26 Dec 2007", "", []goodMailAttachments{}},

{`Received: from 192.168.1.169 (localhost [127.0.0.1])
  by Chris-Lempers-MacBook-Pro.local (Postfix) with ESMTP id 67338820062
  for <MichaelJWilliamstfb24d057-49fb-477d-8cf3-5357f2591641@test.com>; Tue, 27 Mar 2012 13:57:13 -0600 (MDT)
Date: Tue, 27 Mar 2012 13:57:13 -0600 (MDT)
From: support@verical.com
To: MichaelJWilliamstfb24d057-49fb-477d-8cf3-5357f2591641@test.com
Message-ID: <1087061387.3.1332878233226.JavaMail.clemper@Chris-Lempers-MacBook-Pro.local>
Subject: Welcome to Verical
MIME-Version: 1.0
Content-Type: multipart/mixed;
  boundary="----=_Part_0_285260084.1332878232527"

------=_Part_0_285260084.1332878232527
Content-Type: multipart/related;
  boundary="----=_Part_1_1380239564.1332878232593"

------=_Part_1_1380239564.1332878232593
Content-Type: multipart/alternative;
  boundary="----=_Part_2_192678515.1332878232845"

------=_Part_2_192678515.1332878232845
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: 7bit

Please view the HTML version of this email.
------=_Part_2_192678515.1332878232845
Content-Type: text/html;charset=UTF-8
Content-Transfer-Encoding: 7bit

<html>

<head>
</head>

<body>

<table border="0" width="100%" cellspacing="0" cellpadding="0">
<tbody>
<tr bgcolor="#ffffff">
<td align="left">
  <a href="http://www.verical.com/" target="_blank">
    <img border="0" width="155" height="40" alt="Verical.com Logo" src='cid:vericalLogoID'>
  </a>
</td>
</tr>
</tbody>
</table>

<br><br>Dear Michael J. Williams,

<br><br>
<div style="text-align:justify">
Welcome to Verical! Thank you for your registration and for joining the community we're building. Verical is a new type of electronic components distributor, and we're excited to have you as part of this effort. Designed to help firms from throughout the supply chain deal with their inevitable surpluses and shortages, Verical is a one-stop shop for component buyers to purchase the inventory they need straight from the excess inventories of some of the largest firms in the industry.
</div>

<br>Verical is designed to improve the experience of component buyers by delivering several key benefits:
<ul>
    <li><b>Faster search</b> through our dynamic catalog.
    <li><b>Trusted Sources</b> ensured by listing parts drawn exclusively from component manufacturers, franchised distributors, OEM's and contract manufacturers.
    <li><b>Transparent Information</b> like pricing, availability, transit time and detailed part descriptions.
    <li><b>Reliable Fulfillment</b> through our network of world-class third-party logistics providers around the globe.
</ul>

<br>You have been successfully registered with Verical and your account has been activated for use.

<br><br>Log on to <a href="http://localhost:8080/#showLogin=true">http://localhost:8080/#showLogin=true</a> with your email address and password to enter Verical's secure website.

<br><br>Thank you again for joining us, and we look forward to serving you. Please help us improve your experience by filling out a short survey.<br>
<a href="http://www.surveymonkey.com/s/LDNSYZJ">http://www.surveymonkey.com/s/LDNSYZJ</a>
<br><br>

<br><br>Regards,

<br><br>
The entire Verical Team
<br><a href="mailto:clemper@verical.com">clemper@verical.com</a>
<br>1-415-281-3866
<br><br>

<table border="0" width="100%" cellspacing="0" cellpadding="0">
<tbody>
<tr bgcolor="#ffffff">
<td style="font-size: 9px; text-align:justify">
<hr/>
This email may contain information that is legally PRIVILEGED and CONFIDENTIAL intended only for the use of the individual or entity named above.  If you have received this communication in error, please delete all copies of the message and the attachment(s), if any, and promptly notify us at <a href="mailto:clemper@verical.com">clemper@verical.com</a>.  Any unauthorized review, dissemination or copying of this email and its attachments, if any, or the information contained herein is strictly prohibited.
<hr/>
</td>
</tr>
</tbody>
</table>

</body>
</html>
------=_Part_2_192678515.1332878232845--

------=_Part_1_1380239564.1332878232593
Content-Type: image/x-png
Content-Transfer-Encoding: base64
Content-Disposition: inline
Content-ID: <vericalLogoID>

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
jhkFdAMAAQYA7eoayhxF7Y0AAAAASUVORK5CYII=
------=_Part_1_1380239564.1332878232593--

------=_Part_0_285260084.1332878232527--`,
  "Welcome to Verical", "MichaelJWilliamstfb24d057-49fb-477d-8cf3-5357f2591641@test.com", "", "support@verical.com", "", "Please view the HTML version of this email.",
  `<html>

<head>
</head>

<body>

<table border="0" width="100%" cellspacing="0" cellpadding="0">
<tbody>
<tr bgcolor="#ffffff">
<td align="left">
  <a href="http://www.verical.com/" target="_blank">
    <img border="0" width="155" height="40" alt="Verical.com Logo" src='cid:vericalLogoID'>
  </a>
</td>
</tr>
</tbody>
</table>

<br><br>Dear Michael J. Williams,

<br><br>
<div style="text-align:justify">
Welcome to Verical! Thank you for your registration and for joining the community we're building. Verical is a new type of electronic components distributor, and we're excited to have you as part of this effort. Designed to help firms from throughout the supply chain deal with their inevitable surpluses and shortages, Verical is a one-stop shop for component buyers to purchase the inventory they need straight from the excess inventories of some of the largest firms in the industry.
</div>

<br>Verical is designed to improve the experience of component buyers by delivering several key benefits:
<ul>
    <li><b>Faster search</b> through our dynamic catalog.
    <li><b>Trusted Sources</b> ensured by listing parts drawn exclusively from component manufacturers, franchised distributors, OEM's and contract manufacturers.
    <li><b>Transparent Information</b> like pricing, availability, transit time and detailed part descriptions.
    <li><b>Reliable Fulfillment</b> through our network of world-class third-party logistics providers around the globe.
</ul>

<br>You have been successfully registered with Verical and your account has been activated for use.

<br><br>Log on to <a href="http://localhost:8080/#showLogin=true">http://localhost:8080/#showLogin=true</a> with your email address and password to enter Verical's secure website.

<br><br>Thank you again for joining us, and we look forward to serving you. Please help us improve your experience by filling out a short survey.<br>
<a href="http://www.surveymonkey.com/s/LDNSYZJ">http://www.surveymonkey.com/s/LDNSYZJ</a>
<br><br>

<br><br>Regards,

<br><br>
The entire Verical Team
<br><a href="mailto:clemper@verical.com">clemper@verical.com</a>
<br>1-415-281-3866
<br><br>

<table border="0" width="100%" cellspacing="0" cellpadding="0">
<tbody>
<tr bgcolor="#ffffff">
<td style="font-size: 9px; text-align:justify">
<hr/>
This email may contain information that is legally PRIVILEGED and CONFIDENTIAL intended only for the use of the individual or entity named above.  If you have received this communication in error, please delete all copies of the message and the attachment(s), if any, and promptly notify us at <a href="mailto:clemper@verical.com">clemper@verical.com</a>.  Any unauthorized review, dissemination or copying of this email and its attachments, if any, or the information contained herein is strictly prohibited.
<hr/>
</td>
</tr>
</tbody>
</table>

</body>
</html>`, []goodMailAttachments{
  {""},
}},

  {`Date: Mon, 12 Mar 2012 13:13:19 +0000
Return-Path: support@avocosecure.com
To: contactmichaelhart@gmail.com
From: "support@avocosecure.com" <support@avocosecure.com>
Subject: Example subject line
Message-ID: <3bf1b8cb4dca53e79a00931700a8afb0@idplive.local>
X-Priority: 3
X-Mailer: PHPMailer 5.1 (phpmailer.sourceforge.net)
MIME-Version: 1.0
Content-Type: multipart/alternative;
  boundary="b1_3bf1b8cb4dca53e79a00931700a8afb0"


--b1_3bf1b8cb4dca53e79a00931700a8afb0
Content-Type: text/plain; charset = "utf-8"
Content-Transfer-Encoding: 8bit



        Hello
        You have been sent this email as part of your registration with Learner
Passport
        To confirm your email address for use with Learner Passport, please
click on the link below:
        http://example.com/confirm.php
        If clicking on the link does not work, please copy it and paste it into
your browser address entry.
        Thanks,
        Learner Passport, Skills Funding Agency




--b1_3bf1b8cb4dca53e79a00931700a8afb0
Content-Type: text/html; charset = "utf-8"
Content-Transfer-Encoding: 8bit

<!DOCTYPE html>
<html>
<head>
  <style type="text/css">
  p, li {
    font-family: arial;
    line-height: 1.55em;
    margin-bottom: 18px;
  }
  </style>
</head>

<body>
  <table cellspacing='0' cellpadding='0' border='0'>
    <tr>
      <td>&nbsp;</td>
      <td>
        <p>Hello</p>
        <p>You have been sent this email as part of your registration with Learner Passport</p>
        <p>To confirm your email address for use with Learner Passport, please click on the link below:</p>
        http://example.com/confirm.php
        <p>If clicking on the link does not work, please copy it and paste it into your browser address entry.</p>
        <p>Thanks,</p>
        <p>Learner Passport, Skills Funding Agency</p>
      </td>
      <td>&nbsp;</td>
    </tr>
  </table>
</body>
</html>




--b1_3bf1b8cb4dca53e79a00931700a8afb0--`,
  "Example subject line", "contactmichaelhart@gmail.com", "", "support@avocosecure.com", "support@avocosecure.com",
`

        Hello
        You have been sent this email as part of your registration with Learner
Passport
        To confirm your email address for use with Learner Passport, please
click on the link below:
        http://example.com/confirm.php
        If clicking on the link does not work, please copy it and paste it into
your browser address entry.
        Thanks,
        Learner Passport, Skills Funding Agency



`,
`<!DOCTYPE html>
<html>
<head>
  <style type="text/css">
  p, li {
    font-family: arial;
    line-height: 1.55em;
    margin-bottom: 18px;
  }
  </style>
</head>

<body>
  <table cellspacing='0' cellpadding='0' border='0'>
    <tr>
      <td>&nbsp;</td>
      <td>
        <p>Hello</p>
        <p>You have been sent this email as part of your registration with Learner Passport</p>
        <p>To confirm your email address for use with Learner Passport, please click on the link below:</p>
        http://example.com/confirm.php
        <p>If clicking on the link does not work, please copy it and paste it into your browser address entry.</p>
        <p>Thanks,</p>
        <p>Learner Passport, Skills Funding Agency</p>
      </td>
      <td>&nbsp;</td>
    </tr>
  </table>
</body>
</html>



`, []goodMailAttachments{}},

  {`Subject: Hello World
Content-Type: text/plain; charset=ISO-8859-1
Date: Thu, 22 Dec 2011 03:21:05 +0000

ÿôÿý`,
  "Hello World", "", "", "", "", "ÿôÿý", "", []goodMailAttachments{}},

  {`Mime-Version: 1.0 (Apple Message framework v730)
Message-Id: <9169D984-4E0B-45EF-82D4-8F5E53AD7012@example.com>
From: foo@example.com
Subject: testing
Date: Mon, 6 Jun 2005 22:21:22 +0200
To: blah@example.com
Content-Transfer-Encoding: quoted-printable
Content-Type: text/plain

A fax has arrived from remote ID ''.=0D=0A-----------------------=
-------------------------------------=0D=0ATime: 3/9/2006 3:50:52=
 PM=0D=0AReceived from remote ID: =0D=0AInbound user ID XXXXXXXXXX, r=
outing code XXXXXXXXX=0D=0AResult: (0/352;0/0) Successful Send=0D=0AP=
age record: 1 - 1=0D=0AElapsed time: 00:58 on channel 11=0D=0A`,
  "testing", "blah@example.com", "", "foo@example.com", "",
  "A fax has arrived from remote ID ''.=0D=0A-----------------------=\n-------------------------------------=0D=0ATime: 3/9/2006 3:50:52=\n PM=0D=0AReceived from remote ID: =0D=0AInbound user ID XXXXXXXXXX, r=\nouting code XXXXXXXXX=0D=0AResult: (0/352;0/0) Successful Send=0D=0AP=\nage record: 1 - 1=0D=0AElapsed time: 00:58 on channel 11=0D=0A",
  "", []goodMailAttachments{}},

  {`From jamis@37signals.com Mon May  2 16:07:05 2005
Mime-Version: 1.0 (Apple Message framework v622)
Content-Transfer-Encoding: base64
Message-Id: <d3b8cf8e49f04480850c28713a1f473e@37signals.com>
Content-Type: text/plain;
  charset=EUC-KR;
  format=flowed
To: jamis@37signals.com
From: Jamis Buck <jamis@37signals.com>
Subject: Re: Test: =?UTF-8?B?Iua8ouWtlyI=?= mid =?UTF-8?B?Iua8ouWtlyI=?= tail
Date: Mon, 2 May 2005 16:07:05 -0600

tOu6zrrQwMcguLbC+bChwfa3ziwgv+y4rrTCIMfPs6q01MC7ILnPvcC0z7TZLg0KDQrBpiDAzLin
wLogSmFtaXPA1LTPtNku`,
  "Re: Test: \"漢字\" mid \"漢字\" tail", "jamis@37signals.com", "", "jamis@37signals.com", "Jamis Buck", "tOu6zrrQwMcguLbC+bChwfa3ziwgv+y4rrTCIMfPs6q01MC7ILnPvcC0z7TZLg0KDQrBpiDAzLin\nwLogSmFtaXPA1LTPtNku", "", []goodMailAttachments{}},

  {`MIME-Version: 1.0
Subject: =?UTF-8?B?44G+44G/44KA44KB44KC?=
From: Mikel Lindsaar <raasdnil@gmail.com>
To: =?UTF-8?B?44G/44GR44KL?= <raasdnil@gmail.com>
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: base64

44GL44GN44GP44GI44GTCgotLSAKaHR0cDovL2xpbmRzYWFyLm5ldC8KUmFpbHMsIFJTcGVjIGFu
ZCBMaWZlIGJsb2cuLi4uCg==`,
  "まみむめも", "raasdnil@gmail.com", "みける", "raasdnil@gmail.com", "", "44GL44GN44GP44GI44GTCgotLSAKaHR0cDovL2xpbmRzYWFyLm5ldC8KUmFpbHMsIFJTcGVjIGFu\nZCBMaWZlIGJsb2cuLi4uCg==",
  "", []goodMailAttachments{}},

  {`Return-Path: <jeff@37signals.com>
Received: from ?10.0.1.6? (d36-211-30.home1.cgocable.net [24.36.211.30])
        by mx.google.com with ESMTPS id g14sm267889rvb.22.2009.05.13.08.42.03
        (version=TLSv1/SSLv3 cipher=RC4-MD5);
        Wed, 13 May 2009 08:42:04 -0700 (PDT)
Message-Id: <E0F9311D-F469-4E7B-81BC-F240BD473566@37signals.com>
From: Jeffrey Hardy <jeff@37signals.com>
To: Jeffrey Hardy <jeff@37signals.com>
Content-Type: multipart/mixed; boundary=Apple-Mail-6--218366681
Mime-Version: 1.0 (Apple Message framework v935.3)
Subject: =?ISO-8859-1?Q?Eelanal=FC=FCsi_p=E4ring?=
Date: Wed, 13 May 2009 11:42:01 -0400
X-Mailer: Apple Mail (2.935.3)


--Apple-Mail-6--218366681
Content-Disposition: inline;
  filename*=ISO-8859-1''Eelanal%FC%FCsi%20p%E4ring.jpg
Content-Type: image/jpeg;
  x-unix-mode=0700;
  name="=?ISO-8859-1?Q?Eelanal=FC=FCsi_p=E4ring.jpg?="
Content-Transfer-Encoding: base64

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
GwB9/g/9S/t/s9g//9k=

--Apple-Mail-6--218366681--`,
  "Eelanalüüsi päring", "jeff@37signals.com", "Jeffrey Hardy", "jeff@37signals.com", "Jeffrey Hardy", "", "", []goodMailAttachments{
    {"Eelanalüüsi päring.jpg"},
  }},
  {`Subject: this message JUST contains an attachment
From: Ryan Finnie <rfinnie@domain.dom>
To: bob@domain.dom
Content-Disposition: attachment; filename=blah.gz
Content-Transfer-Encoding: base64
Content-Description: Attachment has identical content to above foo.gz
Message-Id: <1066974048.4264.62.camel@localhost>
Mime-Version: 1.0
Date: 23 Oct 2003 22:40:49 -0700
Content-Type: application/x-gzip; NAME=blah.gz

SubjectthismessageJUSTcontainsanattachmentFromRyanFinnierfinniedomaindomTobo
bdomaindomContentDispositionattachmentfilenameAblahgzContentTypeapplication/
xgzipnameAblahgzContentTransferEncodingbase64ContentDescriptionAttachmenthas
identicalcontenttoabovefoogzMessageId1066974048426462camellocalhostMimeVersi
on10Date23Oct20032240490700H4sIAOHBmD8AA4vML1XPyVHISy1LLVJIy8xLUchNVeQCAHbe7
64WA`,
  "this message JUST contains an attachment", "bob@domain.dom", "", "rfinnie@domain.dom", "Ryan Finnie", "", "", []goodMailAttachments{
    {"blah.gz"},
  }},
}
// bad mails

type badMailTypeTest struct {
  RawBody     string
}

var badMailTypeTests = []badMailTypeTest{
  {""},
  {"Invalid email body"},
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


func TestGoodMailParser(t *testing.T) {
  // logger
  log.SetTarget(stdlog.New(os.Stdout, "", stdlog.LstdFlags))
  // uncomment for debug
  //log.Debug = true

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
      expectEq(t, strings.Replace(mail.Text, "\n", "\r\n", -1), email.TextPart, "Value of text")
      expectEq(t, strings.Replace(mail.Html, "\n", "\r\n", -1), email.HtmlPart, "Value of html")
      if len(mail.Attachments) != len(email.Attachments) {
        t.Errorf("Unexpected value for Count of attachments; got %d but expected: %d",
          len(mail.Attachments), len(email.Attachments))
      }
      if len(mail.Attachments) > 0 {
        expectEq(t, mail.Attachments[0].Filename, email.Attachments[0].AttachmentFileName, "Value of attachment name")
      }
    }
  }
}


func TestBadMailParser(t *testing.T) {
  // logger
  log.SetTarget(stdlog.New(os.Stdout, "", stdlog.LstdFlags))
  // uncomment for debug
  //log.Debug = true

  emailParser := EmailParser{}
  for _, mail := range badMailTypeTests {
    testBody := strings.Replace(mail.RawBody, "\n", "\r\n", -1)
    // parse email
    envelop := &smtpd.BasicEnvelope{ MailboxID: 0, MailBody: []byte(testBody)}
    email, err := emailParser.ParseMail(envelop)
    if err == nil {
      t.Error("No error in parsing email: %v", email)
    }
  }
}
