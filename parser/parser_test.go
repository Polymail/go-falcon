package parser

import (
  stdlog "log"
  . "launchpad.net/gocheck"
  "os"
  "encoding/json"
  "strings"
  "testing"
  "math/rand"
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
  "illness", "Mr. X \"wrongquote@b.com\"", "", "sender@mail.com", "Mr. Sender", "  illness 26 Dec - 26 Dec 2007", "", []goodMailAttachments{}},

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
  "Hello World", "", "", "", "", "Ã¿Ã´Ã¿Ã½", "", []goodMailAttachments{}},

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
  "A fax has arrived from remote ID ''.\n------------------------------------------------------------\nTime: 3/9/2006 3:50:52 PM\nReceived from remote ID: \nInbound user ID XXXXXXXXXX, routing code XXXXXXXXX\nResult: (0/352;0/0) Successful Send\nPage record: 1 - 1\nElapsed time: 00:58 on channel 11\n",
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
  "Re: Test: \"漢字\" mid \"漢字\" tail", "jamis@37signals.com", "", "jamis@37signals.com", "Jamis Buck", "대부분의 마찬가지로, 우리는 하나님을 믿습니다.\n\n제 이름은 Jamis입니다.", "", []goodMailAttachments{}},

  {`MIME-Version: 1.0
Subject: =?UTF-8?B?44G+44G/44KA44KB44KC?=
From: Mikel Lindsaar <raasdnil@gmail.com>
To: =?UTF-8?B?44G/44GR44KL?= <raasdnil@gmail.com>
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: base64

44GL44GN44GP44GI44GTCgotLSAKaHR0cDovL2xpbmRzYWFyLm5ldC8KUmFpbHMsIFJTcGVjIGFu
ZCBMaWZlIGJsb2cuLi4uCg==`,
  "まみむめも", "raasdnil@gmail.com", "みける", "raasdnil@gmail.com", "Mikel Lindsaar",
"かきくえこ\n\n-- \nhttp://lindsaar.net/\nRails, RSpec and Life blog....\n",
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

  {`Mime-Version: 1.0 (Apple Message framework v730)
Content-Type: multipart/mixed; boundary=Apple-Mail-13-196941151
Message-Id: <9169D984-4E0B-45EF-82D4-8F5E53AD7012@example.com>
From: foo@example.com
Subject: testing
Date: Mon, 6 Jun 2005 22:21:22 +0200
To: blah@example.com


--Apple-Mail-13-196941151
Content-Transfer-Encoding: quoted-printable
Content-Type: text/plain;
  charset=ISO-8859-1;
  delsp=yes;
  format=flowed

This is the first part.

--Apple-Mail-13-196941151
Content-Type: text/plain; name=This is a test.txt
Content-Transfer-Encoding: 7bit
Content-Disposition: attachment;
  filename=This is a test.txt

Hi there.

--Apple-Mail-13-196941151--`,
  "testing", "blah@example.com", "", "foo@example.com", "", "This is the first part.\n", "", []goodMailAttachments{
    {"This is a test.txt"},
  }},

  {`From xxxxxxxxx.xxxxxxx@gmail.com Sun May  8 19:07:09 2005
Return-Path: <xxxxxxxxx.xxxxxxx@gmail.com>
Message-ID: <e85734b90505081209eaaa17b@mail.gmail.com>
Date: Sun, 8 May 2005 14:09:11 -0500
From: xxxxxxxxx xxxxxxx <xxxxxxxxx.xxxxxxx@gmail.com>
Reply-To: xxxxxxxxx xxxxxxx <xxxxxxxxx.xxxxxxx@gmail.com>
To: xxxxx xxxx <xxxxx@xxxxxxxxx.com>
Subject: Fwd: Signed email causes file attachments
In-Reply-To: <F6E2D0B4-CC35-4A91-BA4C-C7C712B10C13@mac.com>
Mime-Version: 1.0
Content-Type: multipart/mixed;
  boundary="----=_Part_5028_7368284.1115579351471"
References: <F6E2D0B4-CC35-4A91-BA4C-C7C712B10C13@mac.com>

------=_Part_5028_7368284.1115579351471
Content-Type: text/plain; charset=ISO-8859-1
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline

We should not include these files or vcards as attachments.

---------- Forwarded message ----------
From: xxxxx xxxxxx <xxxxxxxx@xxx.com>
Date: May 8, 2005 1:17 PM
Subject: Signed email causes file attachments
To: xxxxxxx@xxxxxxxxxx.com


Hi,

Test attachments oddly encoded with japanese charset.


------=_Part_5028_7368284.1115579351471
Content-Type: application/octet-stream; name*=iso-2022-jp'ja'01%20Quien%20Te%20Dij%8aat.%20Pitbull.mp3
Content-Transfer-Encoding: base64
Content-Disposition: attachment

MIAGCSqGSIb3DQEHAqCAMIACAQExCzAJBgUrDgMCGgUAMIAGCSqGSIb3DQEHAQAAoIIGFDCCAs0w
ggI2oAMCAQICAw5c+TANBgkqhkiG9w0BAQQFADBiMQswCQYDVQQGEwJaQTElMCMGA1UEChMcVGhh
d3RlIENvbnN1bHRpbmcgKFB0eSkgTHRkLjEsMCoGA1UEAxMjVGhhd3RlIFBlcnNvbmFsIEZyZWVt
YWlsIElzc3VpbmcgQ0EwHhcNMDUwMzI5MDkzOTEwWhcNMDYwMzI5MDkzOTEwWjBCMR8wHQYDVQQD
ExZUaGF3dGUgRnJlZW1haWwgTWVtYmVyMR8wHQYJKoZIhvcNAQkBFhBzbWhhdW5jaEBtYWMuY29t
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAn90dPsYS3LjfMY211OSYrDQLzwNYPlAL
7+/0XA+kdy8/rRnyEHFGwhNCDmg0B6pxC7z3xxJD/8GfCd+IYUUNUQV5m9MkxfP9pTVXZVIYLaBw
------=_Part_5028_7368284.1115579351471--`,
  "Fwd: Signed email causes file attachments", "xxxxx@xxxxxxxxx.com", "xxxxx xxxx", "xxxxxxxxx.xxxxxxx@gmail.com", "xxxxxxxxx xxxxxxx",
  `We should not include these files or vcards as attachments.

---------- Forwarded message ----------
From: xxxxx xxxxxx <xxxxxxxx@xxx.com>
Date: May 8, 2005 1:17 PM
Subject: Signed email causes file attachments
To: xxxxxxx@xxxxxxxxxx.com


Hi,

Test attachments oddly encoded with japanese charset.

`, "", []goodMailAttachments{
    {"01 Quien Te Dij\ufffdat. Pitbull.mp3"},
  }},

  {`Date: Thu, 16 May 2013 10:21:11 +0200
From: My Bft <mybft@bft.it>
To: webmaster@bft.it,
 giacomo.macri@develon.com,
 ilenia.trevisan@develon.com
Message-ID: <519496f796057_10c043fd02d0349d012545@mbp_silvia.altavilla.develon.com.mail>
Subject: Bft Oauth development - Export Utenti
Mime-Version: 1.0
Content-Type: multipart/mixed;
 boundary="--==_mimepart_519496f79191c_10c043fd02d0349d012357";
 charset=utf-8
Content-Transfer-Encoding: 7bit



----==_mimepart_519496f79191c_10c043fd02d0349d012357
Date: Thu, 16 May 2013 10:21:11 +0200
Mime-Version: 1.0
Content-Type: text/html;
 charset=utf-8
Content-Transfer-Encoding: 7bit
Content-ID: <519496f793e66_10c043fd02d0349d012428@mbp_silvia.altavilla.develon.com.mail>

<html>
<head>
  <style type="text/css" media="screen">
  a { color: #0077CC; }
  p {
    font-family: Verdana, Arial, Helvetica, Sans-serif;
    font-size: 12px;
    font-weight: normal;
    color: #333333;
    text-align: left;
  }
  #header {
    width: 1000px;
  }
  #footer {
    width: 1000px;
    padding-top: 100px;
  }
  </style>
</head>

<body bgcolor="#FFFFFF" style="margin:0px; padding:0px">
  <table width="100%" border="0" cellspacing="0" cellpadding="0">
    <tr>
      <td width="1000" align="center">
        <div id="header">
          <img alt="Bft-logo-email" src="http://bft-oauth.dev/assets/bft-logo-email.png" style="margin-left: 20px;" />
        </div>
      </td>
    </tr>
  </table>
  <table align="center" width="540" border="0" cellspacing="0" cellpadding="0">
  <tr>
    <td width="540">
      <p style="margin-top:36px"><strong>Nuovo Export Utenti</strong></p>

      <p>
        In Allegato Export Utenti Bft Oauth<br>
      </p>

      <p style="margin-top:40px">Grazie <br/>
        Lo Staff Bft
      </p>

      <p style="padding-top:22px"></p>
    </td>
  </tr>
</table>
  <table align="center" width="1000" border="0" cellspacing="0" cellpadding="0">
    <tr>
      <td width="1000" align="center">
        <img alt="Footer-email" src="http://bft-oauth.dev/assets/footer-email.jpg" />
      </td>
    </tr>
  </table>
</body>

</html>

----==_mimepart_519496f79191c_10c043fd02d0349d012357
Date: Thu, 16 May 2013 10:21:11 +0200
Mime-Version: 1.0
Content-Type: text/csv;
 charset=UTF-8;
 filename=export_users_1368692444.csv
Content-Transfer-Encoding: 7bit
Content-Disposition: attachment;
 filename=export_users_1368692444.csv
Content-ID: <519496f74fd33_10c043fd02d0349d0122c6@mbp_silvia.altavilla.develon.com.mail>

nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,Giacomo,"",giacomo.macri@develon.com,Test,"","","","","","",Italy,"",false,AT
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,ActiveResource,"",activers@develon.com,"",IT03018900245,36077,Vicenza,"",via Retrone 16,0444276203,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,develonlab,develon,lab@develon.com,lab.tes.develon,IT03018900245,36077,Vicenza,Altavilla Vicentina,via Retrone 16,0444276203,Italy,"",false,INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,IGS ELETTRONICA,IACOVELLI,igselettronica@libero.it,IGS ELETTRONICA,02307830592,04022,Latina,Fondi,VIA SAGLIUTOLA 60,3383657585,Italy,"",false,INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,Zilli Gian Domenico,…,zilligiandomenico72@yahoo.it,…,"",32014,Belluno,Ponte nelle Alpi,"Piazza Arrigo Boito, 2/D",0437981420,Italy,Fax. 0437/981420,"",CAT
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,ZAGO ROBERTO ,"",zagoroberto@hotmail.com,ZAGO ROBERTO ,03078270265,31040,Treviso,Salgareda,Via Vigonovo 13,3387418924,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,AUTOMATION SYSTEM,"",info@automationsystem.it,AUTOMATION SYSTEM,02917390235,37040,verona,Veronella,via dell'artigianato,0442480711,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,CALLEGARI MASSIMO,"",masscalle@libero.it,CALLEGARI MASSIMO,03467070268,31030,Treviso,Breda di Piave,cuccilius 1/b,3355232952,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,FABIO,TOFFOLI,toffolifabio@alice.it,TOFFOLI FABIO,01360050932,33080,Pordenone,San Quirino,"VIA ARMENTERESSA, 28",0434919374,Italy,0434919374,true,INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,IVAN,FELETTO,elettrodomus@libero.it,ELETTRODOMUS snc,00422330936,33170,Pordenone,Pordenone,VIA TINTORETTO 7,3483014681,Italy,"",false,INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,ELEKTROIMPIANTI,"",elektroimpiantisnc@vodafone.it,ELEKTROIMPIANTI,02183550306,33034,Udine,Fagagna,VIA DAL BROT 25,3484017931,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,MENGO LANFRANCO,"",mengo.lanfranco@libero.it,MENGO LANFRANCO,00624070272,30016,Venezia,Jesolo,"Via Roma Sx, 23A/7",0421351546,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,LEONARDUZZI SIMONE,"",sissol@alice.it,LEONARDUZZI SIMONE,"","","","","","",Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,Giacomo,riguzzi,riguzzigiacomo@libero.it,ERRE.GI ,03555050404,47521,Forlì-Cesena,Cesena,via Romagna 2700 M.Saraceno,3493257191,Italy,"",false,INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,BET - AM ,"",samuelebettini@tiscali.it,BET - AM ,02233160205,46030,Mantova,Bigarello,"Via Pace, 6",+393498022541,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,ELETTRO 3S,"",info@elettro3s.it,ELETTRO 3S,02426010282,35043,Padova,Monselice,"VIA G.VERDI, 5",0429/784424,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,ELETEL TECHNOLOGY,"",eletel.technology@libero.it,ELETEL TECHNOLOGY,01011860325,34148,Trieste,Trieste,P.ZZA XXV APRILE 7,040812332,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,VENTURINI,"",venturini.impianti@libero.it,VENTURINI,00722400231,37024,Verona,Negrar,Via del Fante 6,337460005,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,UG DI GATTI ULISSE,"",ug.ulisse@virgilio.it,UG DI GATTI ULISSE,"","","","","","",Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,RIGHI IVANO,"",righi.company@virgilio.it,RIGHI IVANO & C snc,00673330353,42040,Reggio Emilia,Campegine,via quartieri  5,0522/677568,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,GIUSEPPE,LOCATELLI,info@gielleelettrica.it,GIELLE ELETTRICA S.R.L.,03130400165,24060,Bergamo,Chiuduno,"VIA A. FANTONI, 22",035838995,Italy,0354496941,false,INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria,CRIVAN,"",ivancri@crivan.191.it,CRIVAN di Crippa Ivan,02437640960,20040,Monza Brianza,Cornate d'Adda,via Fornace,3386273773,Italy,"","",INSTALLER
nome,cognome,email,rag_sociale,p_iva,cap,provincia,comune,indirizzo,telefono,nazione,fax,newsletter,categoria.`,
  "Bft Oauth development - Export Utenti", "webmaster@bft.it", "", "mybft@bft.it", "My Bft", "",
  `<html>
<head>
  <style type="text/css" media="screen">
  a { color: #0077CC; }
  p {
    font-family: Verdana, Arial, Helvetica, Sans-serif;
    font-size: 12px;
    font-weight: normal;
    color: #333333;
    text-align: left;
  }
  #header {
    width: 1000px;
  }
  #footer {
    width: 1000px;
    padding-top: 100px;
  }
  </style>
</head>

<body bgcolor="#FFFFFF" style="margin:0px; padding:0px">
  <table width="100%" border="0" cellspacing="0" cellpadding="0">
    <tr>
      <td width="1000" align="center">
        <div id="header">
          <img alt="Bft-logo-email" src="http://bft-oauth.dev/assets/bft-logo-email.png" style="margin-left: 20px;" />
        </div>
      </td>
    </tr>
  </table>
  <table align="center" width="540" border="0" cellspacing="0" cellpadding="0">
  <tr>
    <td width="540">
      <p style="margin-top:36px"><strong>Nuovo Export Utenti</strong></p>

      <p>
        In Allegato Export Utenti Bft Oauth<br>
      </p>

      <p style="margin-top:40px">Grazie <br/>
        Lo Staff Bft
      </p>

      <p style="padding-top:22px"></p>
    </td>
  </tr>
</table>
  <table align="center" width="1000" border="0" cellspacing="0" cellpadding="0">
    <tr>
      <td width="1000" align="center">
        <img alt="Footer-email" src="http://bft-oauth.dev/assets/footer-email.jpg" />
      </td>
    </tr>
  </table>
</body>

</html>
`, []goodMailAttachments{}},

  {`Message-ID: <1324666081.4ef4cce1a3e1b@dev.immobilier-confiance.fr>
Date: Fri, 23 Dec 2011 19:48:01 +0100
Subject: Alerte suite a la recherche
From: Immobilier Confiance <contact@immobilier-confiance.fr>
To: f.tete@immobilier-confiance.fr
MIME-Version: 1.0
Content-Type: text/html; charset=utf-8
Content-Transfer-Encoding: quoted-printable

Bonjour,
Suite =C3=A0 la recherche ajout=C3=A9e concernant le contact =
Test2 TEST<br/>Voici les r=C3=A9ultats : <br/><br/>R=C3=A9sultats qui peuv=
ent s'accorder aux termes de la recherche :<br/><table><tr><th>R=
=C3=A9f=C3=A9rence</th><th>Type de Bien</th><th>Prix Fai</th><th>N=C3=A9goc=
iateur</th></tr><tr><td>REF901</td><td>ferme</td><td>490000</td><td>oliv=
ier Dal</td></tr><tr><td>REF905</td><td>maison</td><td>269000</td><td>fr=
=C3=A9d=C3=A9ric Ducrot</td></tr><tr><td>REF909</td><td>maison</td><td>234=
000</td><td>olivier Dal</td></tr><tr><td>REF915</td><td>loft</td><td>115=
000</td><td>fr=C3=A9d=C3=A9ric Ducrot</td></tr><tr><td>REF9152</td><td>lof=
t</td><td>125000</td><td>fr=C3=A9d=C3=A9ric Ducrot</td></tr><tr><td>REF927=
</td><td>maison</td><td>179000</td><td>olivier Dal</td></tr></table>`,
  "Alerte suite a la recherche", "f.tete@immobilier-confiance.fr", "", "contact@immobilier-confiance.fr", "Immobilier Confiance", "",
  "Bonjour,\nSuite à la recherche ajoutée concernant le contact Test2 TEST\u003cbr/\u003eVoici les réultats : \u003cbr/\u003e\u003cbr/\u003eRésultats qui peuvent s'accorder aux termes de la recherche :\u003cbr/\u003e\u003ctable\u003e\u003ctr\u003e\u003cth\u003eRéférence\u003c/th\u003e\u003cth\u003eType de Bien\u003c/th\u003e\u003cth\u003ePrix Fai\u003c/th\u003e\u003cth\u003eNégociateur\u003c/th\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF901\u003c/td\u003e\u003ctd\u003eferme\u003c/td\u003e\u003ctd\u003e490000\u003c/td\u003e\u003ctd\u003eolivier Dal\u003c/td\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF905\u003c/td\u003e\u003ctd\u003emaison\u003c/td\u003e\u003ctd\u003e269000\u003c/td\u003e\u003ctd\u003efrédéric Ducrot\u003c/td\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF909\u003c/td\u003e\u003ctd\u003emaison\u003c/td\u003e\u003ctd\u003e234000\u003c/td\u003e\u003ctd\u003eolivier Dal\u003c/td\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF915\u003c/td\u003e\u003ctd\u003eloft\u003c/td\u003e\u003ctd\u003e115000\u003c/td\u003e\u003ctd\u003efrédéric Ducrot\u003c/td\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF9152\u003c/td\u003e\u003ctd\u003eloft\u003c/td\u003e\u003ctd\u003e125000\u003c/td\u003e\u003ctd\u003efrédéric Ducrot\u003c/td\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF927\u003c/td\u003e\u003ctd\u003emaison\u003c/td\u003e\u003ctd\u003e179000\u003c/td\u003e\u003ctd\u003eolivier Dal\u003c/td\u003e\u003c/tr\u003e\u003c/table\u003e",
  []goodMailAttachments{}},


  {`Return-Path: <email_test@me.nowhere>
Received: from omta05sl.mx.bigpond.com by me.nowhere.else with ESMTP id 632BD5758 for <mikel@me.nowhere.else>; Sun, 21 Oct 2007 19:38:21 +1000
Received: from oaamta05sl.mx.bigpond.com by omta05sl.mx.bigpond.com with ESMTP id <20071021093820.HSPC16667.omta05sl.mx.bigpond.com@oaamta05sl.mx.bigpond.com> for <mikel@me.nowhere.else>; Sun, 21 Oct 2007 19:38:20 +1000
Received: from mikel091a by oaamta05sl.mx.bigpond.com with SMTP id <20071021093820.JFMT24025.oaamta05sl.mx.bigpond.com@mikel091a> for <mikel@me.nowhere.else>; Sun, 21 Oct 2007 19:38:20 +1000
Date: Sun, 21 Oct 2007 19:38:13 +1000
From: Mikel Lindsaar <email_test@me.nowhere>
Reply-To: Mikel Lindsaar <email_test@me.nowhere>
To: mikel@me.nowhere
Message-Id: <009601c813c6$19df3510$0437d30a@mikel091a>
Subject: Testing outlook
Mime-Version: 1.0
Content-Type: multipart/alternative; boundary="----=_NextPart_000_0093_01C81419.EB75E850"
X-Get_mail_default: mikel@me.nowhere.else
X-Priority: 3
X-Original-To: mikel@me.nowhere
X-Mailer: Microsoft Outlook Express 6.00.2900.3138
Delivered-To: mikel@me.nowhere
X-Mimeole: Produced By Microsoft MimeOLE V6.00.2900.3138
X-Msmail-Priority: Normal

This is a multi-part message in MIME format.


------=_NextPart_000_0093_01C81419.EB75E850
Content-Type: text/plain; charset=iso-8859-1
Content-Transfer-Encoding: Quoted-printable

Hello
This is an outlook test

So there.

Me.

------=_NextPart_000_0093_01C81419.EB75E850
Content-Type: text/html; charset=iso-8859-1
Content-Transfer-Encoding: Quoted-printable

<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN">
<HTML><HEAD>
<META http-equiv=3DContent-Type content=3D"text/html; =
charset=3Diso-8859-1">
<META content=3D"MSHTML 6.00.6000.16525" name=3DGENERATOR>
<STYLE></STYLE>
</HEAD>
<BODY bgColor=3D#ffffff>
<DIV><FONT face=3DArial size=3D2>Hello</FONT></DIV>
<DIV><FONT face=3DArial size=3D2><STRONG>This is an outlook=20
test</STRONG></FONT></DIV>
<DIV><FONT face=3DArial size=3D2><STRONG></STRONG></FONT>&nbsp;</DIV>
<DIV><FONT face=3DArial size=3D2><STRONG>So there.</STRONG></FONT></DIV>
<DIV><FONT face=3DArial size=3D2></FONT>&nbsp;</DIV>
<DIV><FONT face=3DArial size=3D2>Me.</FONT></DIV></BODY></HTML>


------=_NextPart_000_0093_01C81419.EB75E850--`,
  "Testing outlook", "mikel@me.nowhere", "", "email_test@me.nowhere", "Mikel Lindsaar", "Hello\nThis is an outlook test\n\nSo there.\n\nMe.\n",
  "\u003c!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.0 Transitional//EN\"\u003e\n\u003cHTML\u003e\u003cHEAD\u003e\n\u003cMETA http-equiv=3DContent-Type content=3D\"text/html; =\ncharset=3Diso-8859-1\"\u003e\n\u003cMETA content=3D\"MSHTML 6.00.6000.16525\" name=3DGENERATOR\u003e\n\u003cSTYLE\u003e\u003c/STYLE\u003e\n\u003c/HEAD\u003e\n\u003cBODY bgColor=3D#ffffff\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003eHello\u003c/FONT\u003e\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003e\u003cSTRONG\u003eThis is an outlook=20\ntest\u003c/STRONG\u003e\u003c/FONT\u003e\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003e\u003cSTRONG\u003e\u003c/STRONG\u003e\u003c/FONT\u003e&nbsp;\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003e\u003cSTRONG\u003eSo there.\u003c/STRONG\u003e\u003c/FONT\u003e\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003e\u003c/FONT\u003e&nbsp;\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003eMe.\u003c/FONT\u003e\u003c/DIV\u003e\u003c/BODY\u003e\u003c/HTML\u003e\n\n",
  []goodMailAttachments{}},

  {`From test@example.com  Mon Aug 22 09:45:15 2011
Date: Fri, 19 Aug 2011 10:47:17 +0900
From: Atsushi Yoshida <atsushi@example.com>
Reply-To: rudeboyjet@gmail.com
Subject: Re: TEST
   =?ISO-2022-JP?B?GyRCJUYlOSVIGyhC?=
   =?ISO-2022-JP?B?GyRCJUYlOSVIGyhCJUYlOSVIGyhC?=
To: rudeboyjet@gmail.com
Message-Id: <0CC5E11ED2C1D@example.com>
In-Reply-To: <rid_5582199198@msgid.example.com>
Mime-Version: 1.0
Content-Type: text/plain; charset=iso-2022-jp
Content-Transfer-Encoding: 7bit

Hello`,
  "Re: TEST テスト テスト%F%9%H", "rudeboyjet@gmail.com", "", "atsushi@example.com", "Atsushi Yoshida", "Hello", "", []goodMailAttachments{}},

  {`From: "Doug Sauder" <doug@example.com>
To: "Jürgen Schmürgen" <schmuergen@example.com>
Subject: Die Hasen und die Frösche (Microsoft Outlook 00)
Date: Wed, 17 May 2000 19:11:50 -0400
Message-ID: <NDBBIAKOPKHFGPLCODIGAEKCCHAA.doug@example.com>
MIME-Version: 1.0
Content-Type: text/plain;
	charset="iso-8859-1"
Content-Transfer-Encoding: base64
X-Priority: 3 (Normal)
X-MSMail-Priority: Normal
X-Mailer: Microsoft Outlook IMO, Build 9.0.2416 (9.0.2910.0)
Importance: Normal
X-MimeOLE: Produced By Microsoft MimeOLE V5.00.2314.1300

RGllIEhhc2VuIHVuZCBkaWUgRnL2c2NoZQ0KDQpEaWUgSGFzZW4ga2xhZ3RlbiBlaW5zdCD8YmVy
IGlocmUgbWnfbGljaGUgTGFnZTsgIndpciBsZWJlbiIsIHNwcmFjaCBlaW4gUmVkbmVyLCAiaW4g
c3RldGVyIEZ1cmNodCB2b3IgTWVuc2NoZW4gdW5kIFRpZXJlbiwgZWluZSBCZXV0ZSBkZXIgSHVu
ZGUsIGRlciBBZGxlciwgamEgZmFzdCBhbGxlciBSYXVidGllcmUhIFVuc2VyZSBzdGV0ZSBBbmdz
dCBpc3Qg5HJnZXIgYWxzIGRlciBUb2Qgc2VsYnN0LiBBdWYsIGxh33QgdW5zIGVpbiBm/HIgYWxs
ZW1hbCBzdGVyYmVuLiIgDQoNCkluIGVpbmVtIG5haGVuIFRlaWNoIHdvbGx0ZW4gc2llIHNpY2gg
bnVuIGVyc+R1ZmVuOyBzaWUgZWlsdGVuIGlobSB6dTsgYWxsZWluIGRhcyBhdd9lcm9yZGVudGxp
Y2hlIEdldPZzZSB1bmQgaWhyZSB3dW5kZXJiYXJlIEdlc3RhbHQgZXJzY2hyZWNrdGUgZWluZSBN
ZW5nZSBGcvZzY2hlLCBkaWUgYW0gVWZlciBzYd9lbiwgc28gc2VociwgZGHfIHNpZSBhdWZzIHNj
aG5lbGxzdGUgdW50ZXJ0YXVjaHRlbi4gDQoNCiJIYWx0IiwgcmllZiBudW4gZWJlbiBkaWVzZXIg
U3ByZWNoZXIsICJ3aXIgd29sbGVuIGRhcyBFcnPkdWZlbiBub2NoIGVpbiB3ZW5pZyBhdWZzY2hp
ZWJlbiwgZGVubiBhdWNoIHVucyBm/HJjaHRlbiwgd2llIGlociBzZWh0LCBlaW5pZ2UgVGllcmUs
IHdlbGNoZSBhbHNvIHdvaGwgbm9jaCB1bmds/GNrbGljaGVyIHNlaW4gbfxzc2VuIGFscyB3aXIu
IiANCg==`,
  "Die Hasen und die Frösche (Microsoft Outlook 00)", "schmuergen@example.com", "", "doug@example.com", "Doug Sauder",
  "Die Hasen und die Frösche\n\nDie Hasen klagten einst über ihre mißliche Lage; \"wir leben\", sprach ein Redner, \"in steter Furcht vor Menschen und Tieren, eine Beute der Hunde, der Adler, ja fast aller Raubtiere! Unsere stete Angst ist ärger als der Tod selbst. Auf, laßt uns ein für allemal sterben.\" \n\nIn einem nahen Teich wollten sie sich nun ersäufen; sie eilten ihm zu; allein das außerordentliche Getöse und ihre wunderbare Gestalt erschreckte eine Menge Frösche, die am Ufer saßen, so sehr, daß sie aufs schnellste untertauchten. \n\n\"Halt\", rief nun eben dieser Sprecher, \"wir wollen das Ersäufen noch ein wenig aufschieben, denn auch uns fürchten, wie ihr seht, einige Tiere, welche also wohl noch unglücklicher sein müssen als wir.\" \n",
  "", []goodMailAttachments{}},

  {`From test@example.com  Mon Aug 22 09:45:15 2011
Date: Fri, 19 Aug 2011 10:47:17 +0900
From: Atsushi Yoshida <atsushi@example.com>
Reply-To: rudeboyjet@gmail.com
Subject: Re: TEST
   =?ISO-2022-JP?B?GyRCJUYlOSVIGyhC?=
   =?ISO-2022-JP?B?GyRCJUYlOSVIGyhCJUYlOSVIGyhC?=
To: rudeboyjet@gmail.com
Message-Id: <0CC5E11ED2C1D@example.com>
In-Reply-To: <rid_5582199198@msgid.example.com>
Mime-Version: 1.0
Content-Type: text/plain; charset=iso-2022-jp
Content-Transfer-Encoding: 7bit

Hello`,
  "Re: TEST テスト テスト%F%9%H", "rudeboyjet@gmail.com", "", "atsushi@example.com", "Atsushi Yoshida", "Hello", "", []goodMailAttachments{}},

  {`From: "Doug Sauder" <doug@example.com>
To: "Joe Blow" <jblow@example.com>
Subject: Test message from Microsoft Outlook 00
Date: Wed, 17 May 2000 19:47:24 -0400
Message-ID: <NDBBIAKOPKHFGPLCODIGOEKFCHAA.doug@example.com>
MIME-Version: 1.0
Content-Type: multipart/mixed;
	boundary="----=_NextPart_000_0010_01BFC038.B91BC650"
X-Priority: 3 (Normal)
X-MSMail-Priority: Normal
X-Mailer: Microsoft Outlook IMO, Build 9.0.2416 (9.0.2910.0)
Importance: Normal
X-MimeOLE: Produced By Microsoft MimeOLE V5.00.2314.1300

This is a multi-part message in MIME format.

------=_NextPart_000_0010_01BFC038.B91BC650
Content-Type: multipart/related;
	boundary="----=_NextPart_001_0011_01BFC038.B91BC650"


------=_NextPart_001_0011_01BFC038.B91BC650
Content-Type: multipart/alternative;
	boundary="----=_NextPart_002_0012_01BFC038.B91BC650"


------=_NextPart_002_0012_01BFC038.B91BC650
Content-Type: text/plain;
	charset="iso-8859-1"
Content-Transfer-Encoding: quoted-printable



The Hare and the Tortoise=20
=20
A HARE one day ridiculed the short feet and slow pace of the Tortoise, =
who replied, laughing:  "Though you be swift as the wind, I will beat =
you in a race."  The Hare, believing her assertion to be simply =
impossible, assented to the proposal; and they agreed that the Fox =
should choose the course and fix the goal.  On the day appointed for the =
race the two started together.  The Tortoise never for a moment stopped, =
but went on with a slow but steady pace straight to the end of the =
course.  The Hare, lying down by the wayside, fell fast asleep.  At last =
waking up, and moving as fast as he could, he saw the Tortoise had =
reached the goal, and was comfortably dozing after her fatigue. =20
=20
Slow but steady wins the race. =20



------=_NextPart_002_0012_01BFC038.B91BC650
Content-Type: text/html;
	charset="iso-8859-1"
Content-Transfer-Encoding: quoted-printable

<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN">
<HTML><HEAD>
<META content=3D"text/html; charset=3Diso-8859-1" =
http-equiv=3DContent-Type>
<META content=3D"MSHTML 5.00.2314.1000" name=3DGENERATOR></HEAD>
<BODY>
<DIV><FONT face=3DArial size=3D2><IMG align=3Dbaseline alt=3D"blue ball" =
border=3D0=20
hspace=3D0 src=3D"cid:938014623@17052000-0f9b"></FONT></DIV>
<DIV><FONT face=3DArial size=3D2><BR>The Hare and the Tortoise =
<BR>&nbsp;<BR>A HARE=20
one day ridiculed the short feet and slow pace of the Tortoise, who =
replied,=20
laughing:&nbsp; "Though you be swift as the wind, I will beat you in a=20
race."&nbsp; The Hare, believing her assertion to be simply impossible, =
assented=20
to the proposal; and they agreed that the Fox should choose the course =
and fix=20
the goal.&nbsp; On the day appointed for the race the two started=20
together.&nbsp; The Tortoise never for a moment stopped, but went on =
with a slow=20
but steady pace straight to the end of the course.&nbsp; The Hare, lying =
down by=20
the wayside, fell fast asleep.&nbsp; At last waking up, and moving as =
fast as he=20
could, he saw the Tortoise had reached the goal, and was comfortably =
dozing=20
after her fatigue.&nbsp; <BR>&nbsp;<BR>Slow but steady wins the =
race.&nbsp;=20
</FONT></DIV>
<DIV><FONT face=3DArial size=3D2><BR>&nbsp;</DIV></FONT></BODY></HTML>

------=_NextPart_002_0012_01BFC038.B91BC650--

------=_NextPart_001_0011_01BFC038.B91BC650
Content-Type: image/png;
	name="blueball.png"
Content-Transfer-Encoding: base64
Content-ID: <938014623@17052000-0f9b>

iVBORw0KGgoAAAANSUhEUgAAABsAAAAbCAMAAAC6CgRnAAADAFBMVEX///8AAAgAABAAABgAAAAA
CCkAEEIAEEoACDEAEFIIIXMIKXsIKYQIIWsAGFoACDkIIWMQOZwYQqUYQq0YQrUQOaUQMZQAGFIQ
MYwpUrU5Y8Y5Y84pWs4YSs4YQs4YQr1Ca8Z7nNacvd6Mtd5jlOcxa94hUt4YStYYQsYQMaUAACHO
5+/n7++cxu9ShO8pWucQOa1Ke86tzt6lzu9ajO8QMZxahNat1ufO7++Mve9Ke+8YOaUYSsaMvee1
5++Uve8AAClajOdzpe9rnO8IKYwxY+8pWu8IIXsAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAADBMg1VAAAAAXRSTlMAQObYZgAAABZ0RVh0U29mdHdhcmUAZ2lmMnBuZyAyLjAu
MT1evmgAAAGISURBVHicddJtV5swGAbgEk6AJhBSk4bMCUynBSLaqovbrG/bfPn/vyh70lbssceb
L5xznTsh5BmNhgQoRChwo50EOIohUYLDj4zHhKYQkrEoQdvock4ne0IKMVUpKZLQDeqSTIsv+18P
yqqWUw2IBsRM7307PPp+fDJrWtnpLDJvewYxnewfnvanZ+fzpmwXijC8KbqEa3Fx2ff91Y95U9XC
UpaDeQwiMpHXP/v+1++bWVPWQoGFawtjury9vru/f/C1Vi7ezT0WWpQHf/7+u/G71aLThK/MjRxm
T6KdzZ9fGk9yatMsTgZLl3XVgFRAC6spj/13enssqJVtWVa3NdBSacL8+VZmYqKmdd1CSYoOiMOS
GwtzlqqlFFIuOqv0a1ZEZrUkWICLLFW266y1KvWE1zV/iDAH1EopnVLCiygZCIomH3NCKX0lnI+B
1iuuzCGTxwXjnDO4d7NpbX42YJJHkBwmAm2TxwAZg40J3+Xtbv1rgOAZwG0NxW62p+lT+Yi747sD
/wEUVMzYmWkOvwAAACV0RVh0Q29tbWVudABjbGlwMmdpZiB2LjAuNiBieSBZdmVzIFBpZ3VldDZz
O7wAAAAASUVORK5CYII=

------=_NextPart_001_0011_01BFC038.B91BC650--

------=_NextPart_000_0010_01BFC038.B91BC650
Content-Type: image/png;
	name="greenball.png"
Content-Transfer-Encoding: base64
Content-Disposition: attachment;
	filename="greenball.png"

iVBORw0KGgoAAAANSUhEUgAAABsAAAAbCAMAAAC6CgRnAAADAFBMVEX///8AAAAAEAAAGAAAIQAA
CAAAMQAAQgAAUgAAWgAASgAIYwAIcwAIewAQjAAIawAAOQAAYwAQlAAQnAAhpQAQpQAhrQBCvRhj
xjFjxjlSxiEpzgAYvQAQrQAYrQAhvQCU1mOt1nuE1lJK3hgh1gAYxgAYtQAAKQBCzhDO55Te563G
55SU52NS5yEh3gAYzgBS3iGc52vW75y974yE71JC7xCt73ul3nNa7ykh5wAY1gAx5wBS7yFr7zlK
7xgp5wAp7wAx7wAIhAAQtQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAp1fnZAAAAAXRSTlMAQObYZgAAABZ0RVh0U29mdHdhcmUAZ2lmMnBuZyAyLjAu
MT1evmgAAAFtSURBVHicddJtV8IgFAdwD2zIgMEE1+NcqdsoK+m5tCyz7/+ZiLmHsyzvq53zO/cy
+N9ery1bVe9PWQA9z4MQ+H8Yoj7GASZ95IHfaBGmLOSchyIgyOu22mgQSjUcDuNYcoGjLiLK1cHh
0fHJaTKKOcMItgYxT89OzsfjyTTLC8UF0c2ZNmKquJhczq6ub+YmSVUYRF59GeDastu7+9nD41Nm
kiJ2jc2J3kAWZ9Pr55fH18XSmRuKUTXUaqHy7O19tfr4NFle/w3YDrWRUIlZrL/W86XJkyJVG9Ea
EjIx2XyZmZJGioeUaL+2AY8TY8omR6nkLKhu70zjUKVJXsp3quS2DVSJWNh3zzJKCyexI0ZxBP3a
fE0ElyqOlZJyw8r3BE2SFiJCyxA434SCkg65RhdeQBljQtCg39LWrA90RDDG1EWrYUO23hMANUKR
Rl61E529cR++D2G5LK002dr/qrcfu9u0V3bxn/XdhR/NYeeN0ggsLAAAACV0RVh0Q29tbWVudABj
bGlwMmdpZiB2LjAuNiBieSBZdmVzIFBpZ3VldDZzO7wAAAAASUVORK5CYII=

------=_NextPart_000_0010_01BFC038.B91BC650
Content-Type: image/png;
	name="redball.png"
Content-Transfer-Encoding: base64
Content-Disposition: attachment;
	filename="redball.png"

iVBORw0KGgoAAAANSUhEUgAAABsAAAAbCAMAAAC6CgRnAAADAFBMVEX///8AAAABAAALAAAVAAAa
AAAXAAARAAAKAAADAAAcAAAyAABEAABNAABIAAA9AAAjAAAWAAAmAABhAAB7AACGAACHAAB9AAB0
AABgAAA5AAAUAAAGAAAnAABLAABvAACQAAClAAC7AAC/AACrAAChAACMAABzAABbAAAuAAAIAABM
AAB3AACZAAC0GRnKODjVPT3bKSndBQW4AACoAAB5AAAxAAAYAAAEAABFAACaAAC7JCTRYWHfhITm
f3/mVlbqHx/SAAC5AACjAABdAABCAAAoAAAJAABnAAC6Dw/QVFTek5PlrKzpmZntZWXvJSXXAADB
AACxAACcAABtAABTAAA2AAAbAAAFAABKAACBAADLICDdZ2fonJzrpqbtiorvUVHvFBTRAADDAAC2
AAB4AABeAABAAAAiAABXAACSAADCAADaGxvoVVXseHjveHjvV1fvJibhAADOAAC3AACnAACVAABH
AAArAAAPAACdAADFAADhBQXrKCjvPDzvNTXvGxvjAADQAADJAAC1AACXAACEAABsAABPAAASAAAC
AABiAADpAADvAgLnAADYAADLAAC6AACwAABwAAATAAAkAABYAADIAADTAADNAACzAACDAABuAAAe
AAB+AADAAACkAACNAAB/AABpAABQAAAwAACRAACpAAC8AACqAACbAABlAABJAAAqAAAOAAA0AACs
AACvAACtAACmAACJAAB6AABrAABaAAA+AAApAABqAACCAACfAACeAACWAACPAAB8AAAZAAAHAABV
AACOAACKAAA4AAAQAAA/AAByAACAAABcAAA3AAAsAABmAABDAABWAAAgAAAzAAA8AAA6AAAfAAAM
AAAdAAANAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAD8LtlFAAAAAXRSTlMAQObYZgAAABZ0RVh0U29mdHdhcmUAZ2lmMnBuZyAyLjAu
MT1evmgAAAIISURBVHicY2CAg/8QwIABmJhZWFnZ2Dk4MaU5uLh5eHn5+LkFBDlQJf8zC/EIi4iK
iUtI8koJScsgyf5nlpWTV1BUUlZRVVPX4NFk1UJIyghp6+jq6RsYGhmbKJgK85mZW8Dk/rNaSlhZ
29ja2Ts4Ojkr6Li4urFDNf53N/Ow8vTy9vH18w8IDAoWDQkNC4+ASP5ni4wKio6JjYtPSExKTnFW
SE1LF4A69n9GZlZ2Tm5efkFhUXFySWlZlEd5RSVY7j+TkGRVdU1tXX1DY1Ozcktpa1t7h2YnOAj+
d7l1tyo79vT29SdNSJ44SbFVdHIo9xSIHNPUaWqTpifNSJrZnK00S0U1a/acUG5piNz/uXLzVJ2q
m6dXz584S2WB1cJFi5cshZr539xVftnyFKUVTi2TVjqvyhJLXb1m7TqoHPt6F/HW0g0bN63crGqV
tWXrtu07BJihcsw71+zanRW8Z89eq337RQ/Ip60xO3gIElX/LbikDm8T36KwbNmRo7O3zpHkPSZw
HBqL//8flz1x2OOkyKJTi7aqbzutfUZI2gIuF8F2lr/D5dw2+fZdwpl8YVOlI+CJ4/9/joOyYed5
QzMvhGqnm2V0WiClm///D0lfXHtJ6vLlK9w7rx7vQk5SQJbFtSms1y9evXid7QZacgOxmSxktNzd
tSwwU+J/VICaCPFIYU3XAJhIOtjf5sfyAAAAJXRFWHRDb21tZW50AGNsaXAyZ2lmIHYuMC42IGJ5
IFl2ZXMgUGlndWV0NnM7vAAAAABJRU5ErkJggg==

------=_NextPart_000_0010_01BFC038.B91BC650--
`,
  "Test message from Microsoft Outlook 00", "jblow@example.com", "Joe Blow", "doug@example.com", "Doug Sauder",
  "\n\nThe Hare and the Tortoise \n \nA HARE one day ridiculed the short feet and slow pace of the Tortoise, who replied, laughing:  \"Though you be swift as the wind, I will beat you in a race.\"  The Hare, believing her assertion to be simply impossible, assented to the proposal; and they agreed that the Fox should choose the course and fix the goal.  On the day appointed for the race the two started together.  The Tortoise never for a moment stopped, but went on with a slow but steady pace straight to the end of the course.  The Hare, lying down by the wayside, fell fast asleep.  At last waking up, and moving as fast as he could, he saw the Tortoise had reached the goal, and was comfortably dozing after her fatigue.  \n \nSlow but steady wins the race.  \n\n\n",
  "\u003c!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.0 Transitional//EN\"\u003e\n\u003cHTML\u003e\u003cHEAD\u003e\n\u003cMETA content=\"text/html; charset=iso-8859-1\" http-equiv=Content-Type\u003e\n\u003cMETA content=\"MSHTML 5.00.2314.1000\" name=GENERATOR\u003e\u003c/HEAD\u003e\n\u003cBODY\u003e\n\u003cDIV\u003e\u003cFONT face=Arial size=2\u003e\u003cIMG align=baseline alt=\"blue ball\" border=0 \nhspace=0 src=\"cid:938014623@17052000-0f9b\"\u003e\u003c/FONT\u003e\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=Arial size=2\u003e\u003cBR\u003eThe Hare and the Tortoise \u003cBR\u003e&nbsp;\u003cBR\u003eA HARE \none day ridiculed the short feet and slow pace of the Tortoise, who replied, \nlaughing:&nbsp; \"Though you be swift as the wind, I will beat you in a \nrace.\"&nbsp; The Hare, believing her assertion to be simply impossible, assented \nto the proposal; and they agreed that the Fox should choose the course and fix \nthe goal.&nbsp; On the day appointed for the race the two started \ntogether.&nbsp; The Tortoise never for a moment stopped, but went on with a slow \nbut steady pace straight to the end of the course.&nbsp; The Hare, lying down by \nthe wayside, fell fast asleep.&nbsp; At last waking up, and moving as fast as he \ncould, he saw the Tortoise had reached the goal, and was comfortably dozing \nafter her fatigue.&nbsp; \u003cBR\u003e&nbsp;\u003cBR\u003eSlow but steady wins the race.&nbsp; \n\u003c/FONT\u003e\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=Arial size=2\u003e\u003cBR\u003e&nbsp;\u003c/DIV\u003e\u003c/FONT\u003e\u003c/BODY\u003e\u003c/HTML\u003e\n",
  []goodMailAttachments{
    {"blueball.png"},
    {"greenball.png"},
    {"redball.png"},
  }},

  {`Mime-Version: 1.0 (Apple Message framework v730)
Content-Type: multipart/mixed; boundary=Apple-Mail-13-196941151
Message-Id: <9169D984-4E0B-45EF-82D4-8F5E53AD7012@example.com>
From: foo@example.com
Subject: testing
Date: Mon, 6 Jun 2005 22:21:22 +0200
To: blah@example.com


--Apple-Mail-13-196941151
Content-Transfer-Encoding: quoted-printable
Content-Type: text/plain;
	charset=ISO-8859-1;
	delsp=yes;
	format=flowed

This is the first part.

--Apple-Mail-13-196941151
Content-Type: message/rfc822

From xxxx@xxxx.com Tue May 10 11:28:07 2005
Return-Path: <xxxx@xxxx.com>
X-Original-To: xxxx@xxxx.com
Delivered-To: xxxx@xxxx.com
Received: from localhost (localhost [127.0.0.1])
	by xxx.xxxxx.com (Postfix) with ESMTP id 50FD3A96F
	for <xxxx@xxxx.com>; Tue, 10 May 2005 17:26:50 +0000 (GMT)
Received: from xxx.xxxxx.com ([127.0.0.1])
 by localhost (xxx.xxxxx.com [127.0.0.1]) (amavisd-new, port 10024)
 with LMTP id 70060-03 for <xxxx@xxxx.com>;
 Tue, 10 May 2005 17:26:49 +0000 (GMT)
Received: from xxx.xxxxx.com (xxx.xxxxx.com [69.36.39.150])
	by xxx.xxxxx.com (Postfix) with ESMTP id 8B957A94B
	for <xxxx@xxxx.com>; Tue, 10 May 2005 17:26:48 +0000 (GMT)
Received: from xxx.xxxxx.com (xxx.xxxxx.com [64.233.184.203])
	by xxx.xxxxx.com (Postfix) with ESMTP id 9972514824C
	for <xxxx@xxxx.com>; Tue, 10 May 2005 12:26:40 -0500 (CDT)
Received: by xxx.xxxxx.com with SMTP id 68so1694448wri
        for <xxxx@xxxx.com>; Tue, 10 May 2005 10:26:40 -0700 (PDT)
DomainKey-Signature: a=rsa-sha1; q=dns; c=nofws;
        s=beta; d=xxxxx.com;
        h=received:message-id:date:from:reply-to:to:subject:mime-version:content-type;
        b=g8ZO5ttS6GPEMAz9WxrRk9+9IXBUfQIYsZLL6T88+ECbsXqGIgfGtzJJFn6o9CE3/HMrrIGkN5AisxVFTGXWxWci5YA/7PTVWwPOhJff5BRYQDVNgRKqMl/SMttNrrRElsGJjnD1UyQ/5kQmcBxq2PuZI5Zc47u6CILcuoBcM+A=
Received: by 10.54.96.19 with SMTP id t19mr621017wrb;
        Tue, 10 May 2005 10:26:39 -0700 (PDT)
Received: by 10.54.110.5 with HTTP; Tue, 10 May 2005 10:26:39 -0700 (PDT)
Message-ID: <xxxx@xxxx.com>
Date: Tue, 10 May 2005 11:26:39 -0600
From: Test Tester <xxxx@xxxx.com>
Reply-To: Test Tester <xxxx@xxxx.com>
To: xxxx@xxxx.com, xxxx@xxxx.com
Subject: Another PDF
Mime-Version: 1.0
Content-Type: multipart/mixed;
	boundary="----=_Part_2192_32400445.1115745999735"
X-Virus-Scanned: amavisd-new at textdrive.com

------=_Part_2192_32400445.1115745999735
Content-Type: text/plain; charset=ISO-8859-1
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline

Just attaching another PDF, here, to see what the message looks like,
and to see if I can figure out what is going wrong here.

------=_Part_2192_32400445.1115745999735
Content-Type: application/pdf; name="broken.pdf"
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="broken.pdf"

JVBERi0xLjQNCiXk9tzfDQoxIDAgb2JqDQo8PCAvTGVuZ3RoIDIgMCBSDQogICAvRmlsdGVyIC9G
bGF0ZURlY29kZQ0KPj4NCnN0cmVhbQ0KeJy9Wt2KJbkNvm/od6jrhZxYln9hWEh2p+8HBvICySaE
ycLuTV4/1ifJ9qnq09NpSBimu76yLUuy/qzqcPz7+em3Ixx/CDc6CsXxs3b5+fvfjr/8cPz6/BRu
rbfAx/n3739/fuJylJ5u5fjX81OuDr4deK4Bz3z/aDP+8fz0yw8g0Ofq7ktr1Mn+u28rvhy/jVeD
QSa+9YNKHP/pxjvDNfVAx/m3MFz54FhvTbaseaxiDoN2LeMVMw+yA7RbHSCDzxZuaYB2E1Yay7QU
x89vz0+tyFDKMlAHK5yqLmnjF+c4RjEiQIUeKwblXMe+AsZjN1J5yGQL5DHpDHksurM81rF6PKab
gK6zAarIDzIiUY23rJsN9iorAE816aIu6lsgAdQFsuhhkHOUFgVjp2GjMqSewITXNQ27jrMeamkg
1rPI3iLWG2CIaSBB+V1245YVRICGbbpYKHc2USFDl6M09acQVQYhlwIrkBNLISvXhGlF1wi5FHCw
wxZkoGNJlVeJCEsqKA+3YAV5AMb6KkeaqEJQmFKKQU8T1pRi2ihE1Y4CDrqoYFFXYjJJOatsyzuI
8SIlykuxKTMibWK8H1PgEvqYgs4GmQSrEjJAalgGirIhik+p4ZQN9E3ETFPAHE1b8pp1l/0Rc1gl
fQs0ABWvyoZZzU8VnPXwVVcO9BEsyjEJaO6eBoZRyKGlrKoYoOygA8BGIzgwN3RQ15ouigG5idZQ
fx2U4Db2CqiLO0WHAZoylGiCAqhniNQjFjQPSkmjwfNTgQ6M1Ih+eWo36wFmjIxDJZiGUBiWsAyR
xX3EekGOizkGI96Ol9zVZTAivikURhRsHh2E3JhWMpSTZCnnonrLhMCodgrNcgo4uyJUJc6qnVss
nrGd1Ptr0YwisCOYyIbUwVjV4xBUNLbguSO2YHujonAMJkMdSI7bIw91Akq2AUlMUWGFTMAOamjU
OvZQCxIkY2pCpMFo/IwLdVLHs6nddwTRrgoVbvLU9eB0G4EMndV0TNoxHbt3JBWwK6hhv3iHfDtF
yokB302IpEBTnWICde4uYc/1khDbSIkQopO6lcqamGBu1OSE3N5IPSsZX00CkSHRiiyx6HQIShsS
HSVNswdVsaOUSAWq9aYhDtGDaoG5a3lBGkYt/lFlBFt1UqrYnzVtUpUQnLiZeouKgf1KhRBViRRk
ExepJCzTwEmFDalIRbLEGtw0gfpESOpIAF/NnpPzcVCG86s0g2DuSyd41uhNGbEgaSrWEXORErbw
------=_Part_2192_32400445.1115745999735--

--Apple-Mail-13-196941151--`,
  "testing", "blah@example.com", "", "foo@example.com", "", "This is the first part.\n", "", []goodMailAttachments{
    {"broken.pdf"},
  }},

  {`Date: Mon, 7 Apr 2014 10:35:41 +0000
Return-Path: <info@asdan.org.uk>
To: robforrest@asdan.org.uk
From: ASDAN <info@asdan.org.uk>
Subject: ASDAN password change request
Message-ID: <008e725672a1af9a33a85bddaa6182b0@extranet.asdan.org.uk.local>
X-Priority: 3
X-Mailer: PHPMailer 5.2.7 (https://github.com/PHPMailer/PHPMailer/)
MIME-Version: 1.0
Content-Type: multipart/alternative;
	boundary="b1_008e725672a1af9a33a85bddaa6182b0"
Content-Transfer-Encoding: 8bit

--b1_008e725672a1af9a33a85bddaa6182b0
Content-Type: text/plain; charset=iso-8859-1
Content-Transfer-Encoding: 8bit

This is an empty HTML Snippet that can be edited hereA password reset has been requested for the ASDAN secure area for this email address.
If you did not request this, please delete this email and your password will remain the same.
If you wish to reset your password, please click on the link below where you will be prompted to enter a new password.
Reset your password
If the link above doesn\'t work then please copy and paste the following line(s) into the address bar of your browser including all of the letters and digits.
http://members.asdan.org.uk.local/login/reset_password?hash=b7c97b081249351343f120de0f891247:24389&layout=1
This link will expire in two days.
Kind Regards
ASDAN Centre Support and Training Team
Email : info@asdan.org.uk
Tel : 0117 9411126
Fax : 0117 9351112
This is an empty HTML Snippet that can be edited here




























												About us







									You have received this email either because you have completed an action on our website (asdan.org.uk) or you are subscribed to receive our communications.






										ASDAN Central Office,Wainbrook House, Hudds Vale Road,St George, Bristol BS5 7HY


										t: 0117 941 1126 | f: 0117 935 1112






														info@asdan.org.uk




														www.asdan.org.uk


--b1_008e725672a1af9a33a85bddaa6182b0
Content-Type: text/html; charset=iso-8859-1
Content-Transfer-Encoding: 8bit

<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html>
<head>
	<title>ASDAN</title>
	<meta content="text/html; charset=iso-8859-1" http-equiv="Content-Type"/>
	<style type="text/css">
		* {
			font-family: arial, sans-serif;
		}

		h2 {
			margin: 10px 0 5px 0;
			font-size: 16px;
			font-family: arial, sans-serif;
		}

		p {
			font-size: 14px;
			font-family: arial, sans-serif;
			margin: 5px 0 0 0;
			line-height: 140%;
			color: #424242;
		}

		ul{
			padding-left: 25px;;
			margin: 0;
		}

		li {
			margin: 3px 0 0 0;
			font-size: 14px;
			line-height: 140%;
		}

		hr {
			height: 8px;
			width: 580px;
			padding: 0 10px;
			background: transparent url(http://www.asdan.org.uk/media/images/email/dotted-line-grey-580.png) no-repeat top left;
			border: 0;
		}

		td {
			padding: 0;
			margin: 0;
		}

		td * {
			color: #424242;
		}

		td#page_contents table {
			width: 100%;
			border-collapse: collapse;
			border: solid 1px #dfeaf8;
			padding: 5px 0 0 0;
			margin: 0;
		}

		td#page_contents table td, td#page_contents table th {
			font-size: 14px;
			padding: 5px;
		}

		a { color: #ae3334; text-decoration: none; }

	</style>
</head>
<body marginheight="0" topmargin="0" marginwidth="0" bgcolor="#EEEDEA" leftmargin="0" style="padding: 0; margin: 0;">
<table cellspacing="0" border="0" cellpadding="0" style="margin-top: 0;" width="100%">
	<tr style="background-color: #ffffff; ">
		<td valign="top">
			<table cellspacing="0" border="0" align="center"
			       style="background: #fff; border: 0;" cellpadding="0"
			       width="600">
				<tr>
					<td valign="top">
						<table cellspacing="0" border="0" height="57px" cellpadding="0" width="600">
							<tr>
								<td class="main-title" height="160" valign="top"
								    style="padding: 0; margin:0 ;"
								    width="600">
									<img src=\'http://www.asdan.org.uk/media/images/email/header.png\'
									     width=\'600px\' height=\'160px\'/>
								</td>
							</tr>
						</table>
					</td>
				</tr>
			</table>
		</td>
	</tr>
	<tr style="background-color: #EEEDEA; ">
		<td><img src=\'http://www.asdan.org.uk/media/images/email/spacer-20.gif\'
		         width=\'20px\' height=\'20px\'/></td>
	</tr>
	<tr style="background-color: #EEEDEA; ">
		<td valign="top">
			<table cellspacing="0" border="0" align="center"
			       style="background: #fff;" cellpadding="0"
			       width="600">
				<tr>
					<td>
						<!-- content -->
						<table cellspacing="0" border="0" cellpadding="0" width="600">
							<tr>
								<td style="padding: 0;" valign="top" width="600">
									<table cellspacing="0" border="0" cellpadding="0" width="600px">
										<tr><td colspan="3" width="600px"><img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
										                         width=\'600px\' height=\'10px\' /></td></tr>
										<tr>
											<td width="10px">
												<img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
											                      width=\'10px\' height=\'10px\' /></td>
											<td width="580px" id="page_contents">
												This is an empty HTML Snippet that can be edited <a style=\'color: #ae3334;\'  href=\'http://cms.webdev.asdan.org.uk/content/html_snippets/html_snippet/579\' target=\'_blank\'>here</a><p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >A password reset has been requested for the <a style=\'color: #ae3334;\'  href=\'http://asdan.org.uk\'>ASDAN</a> secure area for this email address.</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >If you did not request this, please delete this email and your password will remain the same.</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >If you wish to reset your password, please click on the link below where you will be prompted to enter a new password.</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' ><a style=\'color: #ae3334;\'  href=\'http://members.asdan.org.uk.local/login/reset_password?hash=b7c97b081249351343f120de0f891247:24389&layout=1\'>Reset your password</a></p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >If the link above doesn\'t work then please copy and paste the following line(s) into the address bar of your browser including all of the letters and digits.</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' ><span style="color: #ae3334;">http://members.asdan.org.uk.local/login/reset_password?hash=b7c97b081249351343f120de0f891247:24389&layout=1</span></p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >This link will expire in two days.</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >Kind Regards</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >ASDAN Centre Support and Training Team</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >Email : <a style=\'color: #ae3334;\'  href=\'mailto:info@asdan.org.uk\'>info@asdan.org.uk</a><br />
Tel : 0117 9411126<br />
Fax : 0117 9351112</p>
This is an empty HTML Snippet that can be edited <a style=\'color: #ae3334;\'  href=\'http://cms.webdev.asdan.org.uk/content/html_snippets/html_snippet/580\' target=\'_blank\'>here</a>											</td>
											<td width="10px">
												<img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
												     width=\'10px\' height=\'10px\' /></td>
										</tr>
										<tr><td colspan="3" width="600px"><img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
										                                       width=\'600px\' height=\'10px\' /></td></tr>
									</table>
								</td>

							</tr>
						</table>
					</td>
				</tr>
			</table>
		</td>
	</tr>
	<tr style="background-color: #EEEDEA; ">
		<td><img src=\'http://www.asdan.org.uk/media/images/email/spacer-20.gif\'
		         width=\'20px\' height=\'20px\'/></td>
	</tr>
	<tr style="background-color: #C2C3C1;">
		<td valign="top">
			<table cellspacing="0" border="0" align="center"
			       style="border: 0;" cellpadding="0"
			       width="600">
				<tr>
					<td valign="top">
						<table cellspacing="0" border="0" cellpadding="0" width="600px" height="130px">
							<tr>
								<td width="400px" valign="top">
									<table width="400px" cellspacing="0" border="0" cellpadding="0">
										<tr>
											<td valign="top"
											    style="background-color: #333333; color: white; font-size: 15px; padding: 10px; font-family: arial,sans-serif; margin-top: 0; display:inline-block; ">
												About us
											</td>
										</tr>
										<tr>
											<td></td>
										</tr>
									</table>

									<p><span style="font-size:10px"><span style="font-family:arial,helvetica,sans-serif">You have received this email either because you have completed an action on our website (<a href="http://www.asdan.org.uk" style=\'text-decoration:none\'><span style="color:#ae3334">asdan.org.uk</span></a>) or you are subscribed to receive our communications.</span></span></p>
								</td>
								<td width="10px">
									<img src=\'http://www.asdan.org.uk/media/images/email/vertical-line-100.png\'
									     width=\'10px\' height=\'100px\'/>
								</td>
								<td width="190px" valign="top">
									<p style="margin: 10px 0 0 0 ; font-family:arial,helvetica,sans-serif;font-size:10px">
										ASDAN Central Office,<br />Wainbrook House, Hudds Vale Road,<br />St George, Bristol BS5 7HY
									</p>
									<p style="margin: 5px 0 0 0 ; font-family:arial,helvetica,sans-serif; font-size:10px">
										t: 0117 941 1126 | f: 0117 935 1112
									</p>
									<table cellspacing="0" border="0" cellpadding="0" width="190px" style="margin-top: 5px; border: 0;">
										<tr>
											<td width="90px" valign="top">
												<p style="margin: 0 ; font-family:arial,helvetica,sans-serif; font-size:10px">
													<a style="text-decoration: none; color: #ae3334;" href="http://mailto:info@asdan.org.uk">
														info@asdan.org.uk
													</a>
												</p>
												<p style="margin: 5px 0 0 0; font-family:arial,helvetica,sans-serif; font-size:10px">
													<a style="text-decoration: none; color: #ae3334;" href="http://www.asdan.org.uk">
														www.asdan.org.uk &pound; or &#163;
													</a>
												</p>
											</td>
											<td width="10px"><img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
											         width=\'10px\' height=\'1px\'/></td>
											<td width="70px" valign="top" align="right">
												<table width="70px" cellspacing="0" border="0" cellpadding="0">
													<tr>
														<td>
															<a href="https://www.facebook.com/pages/ASDAN-education/201528503367262" style="border: 0"><img src=\'http://www.asdan.org.uk/media/images/email/facebook.png\'
															                                                                                                width=\'30px\' height=\'30px\' style="border: 0" /></td>
														<td><img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
														         width=\'10px\' height=\'1px\'/></td>
														<td>
															<a href="http://www.twitter.com/ASDANeducation" style="border: 0"><img src=\'http://www.asdan.org.uk/media/images/email/twitter.png\'
															                                                                       width=\'30px\' height=\'30px\' /></td>
													</tr>
													<tr>
														<td colspan="3"><img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
														                     width=\'10px\' height=\'1px\'/></td>
													</tr>
												</table>
											</td>
										</tr>
									</table>

								</td>
							</tr>
						</table>
					</td>
				</tr>
			</table>
		</td>

	</tr>
</table>
</body>
</html>



--b1_008e725672a1af9a33a85bddaa6182b0--`,
  "ASDAN password change request", "robforrest@asdan.org.uk", "", "info@asdan.org.uk", "ASDAN", `This is an empty HTML Snippet that can be edited hereA password reset has been requested for the ASDAN secure area for this email address.
If you did not request this, please delete this email and your password will remain the same.
If you wish to reset your password, please click on the link below where you will be prompted to enter a new password.
Reset your password
If the link above doesn\'t work then please copy and paste the following line(s) into the address bar of your browser including all of the letters and digits.
http://members.asdan.org.uk.local/login/reset_password?hash=b7c97b081249351343f120de0f891247:24389&layout=1
This link will expire in two days.
Kind Regards
ASDAN Centre Support and Training Team
Email : info@asdan.org.uk
Tel : 0117 9411126
Fax : 0117 9351112
This is an empty HTML Snippet that can be edited here




























												About us







									You have received this email either because you have completed an action on our website (asdan.org.uk) or you are subscribed to receive our communications.






										ASDAN Central Office,Wainbrook House, Hudds Vale Road,St George, Bristol BS5 7HY


										t: 0117 941 1126 | f: 0117 935 1112






														info@asdan.org.uk




														www.asdan.org.uk

`, `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html>
<head>
	<title>ASDAN</title>
	<meta content="text/html; charset=iso-8859-1" http-equiv="Content-Type"/>
	<style type="text/css">
		* {
			font-family: arial, sans-serif;
		}

		h2 {
			margin: 10px 0 5px 0;
			font-size: 16px;
			font-family: arial, sans-serif;
		}

		p {
			font-size: 14px;
			font-family: arial, sans-serif;
			margin: 5px 0 0 0;
			line-height: 140%;
			color: #424242;
		}

		ul{
			padding-left: 25px;;
			margin: 0;
		}

		li {
			margin: 3px 0 0 0;
			font-size: 14px;
			line-height: 140%;
		}

		hr {
			height: 8px;
			width: 580px;
			padding: 0 10px;
			background: transparent url(http://www.asdan.org.uk/media/images/email/dotted-line-grey-580.png) no-repeat top left;
			border: 0;
		}

		td {
			padding: 0;
			margin: 0;
		}

		td * {
			color: #424242;
		}

		td#page_contents table {
			width: 100%;
			border-collapse: collapse;
			border: solid 1px #dfeaf8;
			padding: 5px 0 0 0;
			margin: 0;
		}

		td#page_contents table td, td#page_contents table th {
			font-size: 14px;
			padding: 5px;
		}

		a { color: #ae3334; text-decoration: none; }

	</style>
</head>
<body marginheight="0" topmargin="0" marginwidth="0" bgcolor="#EEEDEA" leftmargin="0" style="padding: 0; margin: 0;">
<table cellspacing="0" border="0" cellpadding="0" style="margin-top: 0;" width="100%">
	<tr style="background-color: #ffffff; ">
		<td valign="top">
			<table cellspacing="0" border="0" align="center"
			       style="background: #fff; border: 0;" cellpadding="0"
			       width="600">
				<tr>
					<td valign="top">
						<table cellspacing="0" border="0" height="57px" cellpadding="0" width="600">
							<tr>
								<td class="main-title" height="160" valign="top"
								    style="padding: 0; margin:0 ;"
								    width="600">
									<img src=\'http://www.asdan.org.uk/media/images/email/header.png\'
									     width=\'600px\' height=\'160px\'/>
								</td>
							</tr>
						</table>
					</td>
				</tr>
			</table>
		</td>
	</tr>
	<tr style="background-color: #EEEDEA; ">
		<td><img src=\'http://www.asdan.org.uk/media/images/email/spacer-20.gif\'
		         width=\'20px\' height=\'20px\'/></td>
	</tr>
	<tr style="background-color: #EEEDEA; ">
		<td valign="top">
			<table cellspacing="0" border="0" align="center"
			       style="background: #fff;" cellpadding="0"
			       width="600">
				<tr>
					<td>
						<!-- content -->
						<table cellspacing="0" border="0" cellpadding="0" width="600">
							<tr>
								<td style="padding: 0;" valign="top" width="600">
									<table cellspacing="0" border="0" cellpadding="0" width="600px">
										<tr><td colspan="3" width="600px"><img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
										                         width=\'600px\' height=\'10px\' /></td></tr>
										<tr>
											<td width="10px">
												<img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
											                      width=\'10px\' height=\'10px\' /></td>
											<td width="580px" id="page_contents">
												This is an empty HTML Snippet that can be edited <a style=\'color: #ae3334;\'  href=\'http://cms.webdev.asdan.org.uk/content/html_snippets/html_snippet/579\' target=\'_blank\'>here</a><p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >A password reset has been requested for the <a style=\'color: #ae3334;\'  href=\'http://asdan.org.uk\'>ASDAN</a> secure area for this email address.</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >If you did not request this, please delete this email and your password will remain the same.</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >If you wish to reset your password, please click on the link below where you will be prompted to enter a new password.</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' ><a style=\'color: #ae3334;\'  href=\'http://members.asdan.org.uk.local/login/reset_password?hash=b7c97b081249351343f120de0f891247:24389&layout=1\'>Reset your password</a></p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >If the link above doesn\'t work then please copy and paste the following line(s) into the address bar of your browser including all of the letters and digits.</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' ><span style="color: #ae3334;">http://members.asdan.org.uk.local/login/reset_password?hash=b7c97b081249351343f120de0f891247:24389&layout=1</span></p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >This link will expire in two days.</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >Kind Regards</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >ASDAN Centre Support and Training Team</p>
<p style=\'line-height: 140%; color: #424242; font-size: 14px; margin: 5px 0 0 0; font-family: arial,sans-serif;\' >Email : <a style=\'color: #ae3334;\'  href=\'mailto:info@asdan.org.uk\'>info@asdan.org.uk</a><br />
Tel : 0117 9411126<br />
Fax : 0117 9351112</p>
This is an empty HTML Snippet that can be edited <a style=\'color: #ae3334;\'  href=\'http://cms.webdev.asdan.org.uk/content/html_snippets/html_snippet/580\' target=\'_blank\'>here</a>											</td>
											<td width="10px">
												<img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
												     width=\'10px\' height=\'10px\' /></td>
										</tr>
										<tr><td colspan="3" width="600px"><img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
										                                       width=\'600px\' height=\'10px\' /></td></tr>
									</table>
								</td>

							</tr>
						</table>
					</td>
				</tr>
			</table>
		</td>
	</tr>
	<tr style="background-color: #EEEDEA; ">
		<td><img src=\'http://www.asdan.org.uk/media/images/email/spacer-20.gif\'
		         width=\'20px\' height=\'20px\'/></td>
	</tr>
	<tr style="background-color: #C2C3C1;">
		<td valign="top">
			<table cellspacing="0" border="0" align="center"
			       style="border: 0;" cellpadding="0"
			       width="600">
				<tr>
					<td valign="top">
						<table cellspacing="0" border="0" cellpadding="0" width="600px" height="130px">
							<tr>
								<td width="400px" valign="top">
									<table width="400px" cellspacing="0" border="0" cellpadding="0">
										<tr>
											<td valign="top"
											    style="background-color: #333333; color: white; font-size: 15px; padding: 10px; font-family: arial,sans-serif; margin-top: 0; display:inline-block; ">
												About us
											</td>
										</tr>
										<tr>
											<td></td>
										</tr>
									</table>

									<p><span style="font-size:10px"><span style="font-family:arial,helvetica,sans-serif">You have received this email either because you have completed an action on our website (<a href="http://www.asdan.org.uk" style=\'text-decoration:none\'><span style="color:#ae3334">asdan.org.uk</span></a>) or you are subscribed to receive our communications.</span></span></p>
								</td>
								<td width="10px">
									<img src=\'http://www.asdan.org.uk/media/images/email/vertical-line-100.png\'
									     width=\'10px\' height=\'100px\'/>
								</td>
								<td width="190px" valign="top">
									<p style="margin: 10px 0 0 0 ; font-family:arial,helvetica,sans-serif;font-size:10px">
										ASDAN Central Office,<br />Wainbrook House, Hudds Vale Road,<br />St George, Bristol BS5 7HY
									</p>
									<p style="margin: 5px 0 0 0 ; font-family:arial,helvetica,sans-serif; font-size:10px">
										t: 0117 941 1126 | f: 0117 935 1112
									</p>
									<table cellspacing="0" border="0" cellpadding="0" width="190px" style="margin-top: 5px; border: 0;">
										<tr>
											<td width="90px" valign="top">
												<p style="margin: 0 ; font-family:arial,helvetica,sans-serif; font-size:10px">
													<a style="text-decoration: none; color: #ae3334;" href="http://mailto:info@asdan.org.uk">
														info@asdan.org.uk
													</a>
												</p>
												<p style="margin: 5px 0 0 0; font-family:arial,helvetica,sans-serif; font-size:10px">
													<a style="text-decoration: none; color: #ae3334;" href="http://www.asdan.org.uk">
														www.asdan.org.uk &pound; or &#163;
													</a>
												</p>
											</td>
											<td width="10px"><img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
											         width=\'10px\' height=\'1px\'/></td>
											<td width="70px" valign="top" align="right">
												<table width="70px" cellspacing="0" border="0" cellpadding="0">
													<tr>
														<td>
															<a href="https://www.facebook.com/pages/ASDAN-education/201528503367262" style="border: 0"><img src=\'http://www.asdan.org.uk/media/images/email/facebook.png\'
															                                                                                                width=\'30px\' height=\'30px\' style="border: 0" /></td>
														<td><img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
														         width=\'10px\' height=\'1px\'/></td>
														<td>
															<a href="http://www.twitter.com/ASDANeducation" style="border: 0"><img src=\'http://www.asdan.org.uk/media/images/email/twitter.png\'
															                                                                       width=\'30px\' height=\'30px\' /></td>
													</tr>
													<tr>
														<td colspan="3"><img src=\'http://www.asdan.org.uk/media/images/email/spacer-10.gif\'
														                     width=\'10px\' height=\'1px\'/></td>
													</tr>
												</table>
											</td>
										</tr>
									</table>

								</td>
							</tr>
						</table>
					</td>
				</tr>
			</table>
		</td>

	</tr>
</table>
</body>
</html>


`, []goodMailAttachments{}},

  {`Return-Path: <mail_dump@ns1.sssssss.net.au>
Received: from acomputer ([unix socket])
	 by imap01.sssssss.net (Cyrus) with LMTPA;
	 Wed, 16 Jan 2008 13:51:42 -0800
X-Sieve: CMU Sieve 2.2
Received: from smtp.sssssss.org (unknown [198.0.0.92])
	by imap01.sssssss.net (Postfix) with ESMTP id BFE6477FAE
	for <aaaa@sssssss.net>; Wed, 16 Jan 2008 13:51:40 -0800 (PST)
Received: from ns1.sssssss.net.au (ns1.sssssss.net.au [202.0.0.246])
	by smtp.sssssss.org (Postfix) with ESMTP id 96E5C6C6830
	for <aaaa@sssssss.net>; Wed, 16 Jan 2008 11:08:14 -0800 (PST)
Received: from ns1.sssssss.net.au (unknown [127.0.0.1])
	by ns1.sssssss.net.au (Postfix) with ESMTP id E1BCEF227
	for <aaaa@sssssss.net>; Thu, 17 Jan 2008 03:40:53 +1100 (EST)
Received: from ns1.sssssss.net.au (ns1.sssssss.net.au [202.0.0.246])
        by localhost (FormatMessage) with SMTP id ceaa681bbcb6c7f6
        for <jennifer@sss.sssssss.net.au>; Thu, 17 Jan 2008 03:40:53 +1100 (EST)
Received: from mail11.tppppp.com.au (unknown [203.0.0.161])
	by ns1.sssssss.net.au (Postfix) with ESMTP id 7F2D2F225
	for <jennifer@sss.sssssss.net.au>; Thu, 17 Jan 2008 03:40:52 +1100 (EST)
Received: from localhost (localhost)
	by mail11.tppppp.com.au (envelope-from MAILER-DAEMON) (8.14.2/8.14.2) id m0GFZ1c3009410;
	Thu, 17 Jan 2008 03:40:52 +1100
Date: Thu, 17 Jan 2008 03:40:52 +1100
From: Mail Delivery Subsystem <MAILER-DAEMON@tppppp.com.au>
Message-Id: <200801161640.m0GFZ1c3009410@mail11.ttttt.com.au>
To: <jennifer@sss.sssssss.net.au>
MIME-Version: 1.0
Content-Type: multipart/report; report-type=delivery-status;
	boundary="m0GFZ1c3009410.1200501652/mail11.ttttt.com.au"
Subject: Warning: could not send message for past 8 hours
Auto-Submitted: auto-generated (warning-timeout)
Resent-Date: Thu, 17 Jan 2008 03:40:53 +1100 (EST)
Resent-From: <mail_dump@ns1.sssssss.net.au>
Resent-To: <mikel@sssssss.net>
Resent-Message-ID: <ceaa681bbcb6c7f6.1200501653@sssssss.net.au>
X-Spam-Status: No


--m0GFZ1c3009410.1200501652/mail11.ttttt.com.au
Content-Type: text/plain

    **********************************************
    **      THIS IS A WARNING MESSAGE ONLY      **
    **  YOU DO NOT NEED TO RESEND YOUR MESSAGE  **
    **********************************************

The original message was received at Wed, 16 Jan 2008 19:38:07 +1100
from 60-0-0-61.static.tppppp.com.au [60.0.0.61]

This message was generated by mail11.tppppp.com.au

   ----- Transcript of session follows -----
.... while talking to mail.oooooooo.com.au.:
>>> DATA
<<< 452 4.2.2 <fraser@oooooooo.com.au>... Mailbox full
<fraser@oooooooo.com.au>... Deferred: 452 4.2.2 <fraser@oooooooo.com.au>... Mailbox full
<<< 503 5.0.0 Need RCPT (recipient)
Warning: message still undelivered after 8 hours
Will keep trying until message is 5 days old

--
This message has been scanned for viruses and
dangerous content by MailScanner, and is
believed to be clean.


--m0GFZ1c3009410.1200501652/mail11.ttttt.com.au
Content-Type: message/delivery-status

Reporting-MTA: dns; mail11.ttttt.com.au
Arrival-Date: Wed, 16 Jan 2008 19:38:07 +1100

Final-Recipient: RFC822; fraser@oooooooo.com.au
Action: delayed
Status: 4.2.2
Remote-MTA: DNS; mail.oooooooo.com.au
Diagnostic-Code: SMTP; 452 4.2.2 <fraser@oooooooo.com.au>... Mailbox full
Last-Attempt-Date: Thu, 17 Jan 2008 03:40:52 +1100

--m0GFZ1c3009410.1200501652/mail11.ttttt.com.au
Content-Type: text/rfc822-headers

Return-Path: <jennifer@sss.sssssss.net.au>
Received: from k1s2yo86 (60-0-0-61.static.tppppp.com.au [60.0.0.61])
	by mail11.tppppp.com.au (envelope-from jennifer@sss.sssssss.net.au) (8.14.2/8.14.2) with ESMTP id m0G8c0fR020461
	for <fraser@oooooooo.com.au>; Wed, 16 Jan 2008 19:38:07 +1100
Date: Wed, 16 Jan 2008 19:38:07 +1100
From: Sydney <jennifer@sss.sssssss.net.au>
Message-ID: <15655788.13.1200472642578.JavaMail.Administrator@mail.ttttt.com.au>
Subject: Wanted
MIME-Version: 1.0
Content-Type: multipart/mixed;
	boundary="----=_Part_12_28168925.1200472642578"
X-Virus-Scanned: ClamAV 0.91.2/5484/Wed Jan 16 06:31:27 2008 on mail11.tppppp.com.au
X-Virus-Status: Clean

--m0GFZ1c3009410.1200501652/mail11.ttttt.com.au--`,
  "Warning: could not send message for past 8 hours", "jennifer@sss.sssssss.net.au", "", "MAILER-DAEMON@tppppp.com.au", "Mail Delivery Subsystem", `    **********************************************
    **      THIS IS A WARNING MESSAGE ONLY      **
    **  YOU DO NOT NEED TO RESEND YOUR MESSAGE  **
    **********************************************

The original message was received at Wed, 16 Jan 2008 19:38:07 +1100
from 60-0-0-61.static.tppppp.com.au [60.0.0.61]

This message was generated by mail11.tppppp.com.au

   ----- Transcript of session follows -----
.... while talking to mail.oooooooo.com.au.:
>>> DATA
<<< 452 4.2.2 <fraser@oooooooo.com.au>... Mailbox full
<fraser@oooooooo.com.au>... Deferred: 452 4.2.2 <fraser@oooooooo.com.au>... Mailbox full
<<< 503 5.0.0 Need RCPT (recipient)
Warning: message still undelivered after 8 hours
Will keep trying until message is 5 days old

--
This message has been scanned for viruses and
dangerous content by MailScanner, and is
believed to be clean.

Reporting-MTA: dns; mail11.ttttt.com.au
Arrival-Date: Wed, 16 Jan 2008 19:38:07 +1100

Final-Recipient: RFC822; fraser@oooooooo.com.au
Action: delayed
Status: 4.2.2
Remote-MTA: DNS; mail.oooooooo.com.au
Diagnostic-Code: SMTP; 452 4.2.2 <fraser@oooooooo.com.au>... Mailbox full
Last-Attempt-Date: Thu, 17 Jan 2008 03:40:52 +1100
Return-Path: <jennifer@sss.sssssss.net.au>
Received: from k1s2yo86 (60-0-0-61.static.tppppp.com.au [60.0.0.61])
	by mail11.tppppp.com.au (envelope-from jennifer@sss.sssssss.net.au) (8.14.2/8.14.2) with ESMTP id m0G8c0fR020461
	for <fraser@oooooooo.com.au>; Wed, 16 Jan 2008 19:38:07 +1100
Date: Wed, 16 Jan 2008 19:38:07 +1100
From: Sydney <jennifer@sss.sssssss.net.au>
Message-ID: <15655788.13.1200472642578.JavaMail.Administrator@mail.ttttt.com.au>
Subject: Wanted
MIME-Version: 1.0
Content-Type: multipart/mixed;
	boundary="----=_Part_12_28168925.1200472642578"
X-Virus-Scanned: ClamAV 0.91.2/5484/Wed Jan 16 06:31:27 2008 on mail11.tppppp.com.au
X-Virus-Status: Clean
`, "", []goodMailAttachments{}},

  {`Return-Path: <>
X-Original-To: notification+promo@blah.com
Delivered-To: notification+promo@blah.com
Received: from schemailmta04.ci.com (schemailmta04.ci.com [209.183.37.58])
    by blah.com (Postfix) with ESMTP id 24EF419F546
    for <notification+promo@blah.com>; Tue, 29 Jun 2010 15:42:46 +0000 (UTC)
To: notification+promo@blah.com
From: Mail Administrator <Postmaster@ci.com>
Reply-To: <Postmaster@ci.com>
Subject: Mail System Error - Returned Mail
Date: Tue, 29 Jun 2010 10:42:44 -0500
Message-ID: <20100629154244.OZPA15102.schemailmta04.ci.com@schemailmta04>
MIME-Version: 1.0
Content-Type: multipart/report;
        report-type=delivery-status;
        Boundary="===========================_ _= 6078796(15102)1277826164"
X-Cloudmark-Analysis: v=1.0 c=1 a=q8OS1GolVHwA:10 a=ev1gGZlfZ-EA:10 a=HQ-Cukr2AAAA:8 a=qihIh-XuXL65y3o_mUgA:9 a=mUL5bUDOV_-gjcCZylcY5Lz4jjsA:4 a=iQvSWfByulMA:10 a=ni8l3qMSI1sA:10 a=WHDNLAQ519cA:10 a=Fry9e7MVxuJdODrS104A:9 a=JYo4OF_E9TqbHrUN2TvLdggtx2cA:4 a=S0jCPnXDAAAA:8 a=pXkHMj1YAAAA:8 a=5ErcFzC0N3E7OloTRA8A:9 a=cC0RL7HlXt3RrKfnpEbxHCeM-zQA:4 a=cHEBK1Z0Lu8A:10 a=p9ZeupWRHUwA:10 a=7sPVfr_AX1EA:10


--===========================_ _= 6078796(15102)1277826164
Content-Type: message/delivery-status;

This Message was undeliverable due to the following reason:


<u@ci.com> has restricted SMS e-mail

Please reply to <Postmaster@ci.com>
if you feel this message to be in error.

--===========================_ _= 6078796(15102)1277826164
Content-Type: message/delivery-status

Reporting-MTA: dns; schemailmta04.ci.com
Arrival-Date: Tue, 29 Jun 2010 10:42:37 -0500
Received-From-MTA: dns; schemailedgegx04.ci.com (172.16.130.170)

Original-Recipient: rfc822;u@ci.com
Final-Recipient: RFC822; <u@ci.com>
Action: failed
Status: 5.3.0

--===========================_ _= 6078796(15102)1277826164
Content-Type: message/rfc822

Received: from schemailedgegx04.ci.com ([172.16.130.170])
          by schemailmta04.ci.com
          (InterMail vM.6.01.04.00 201-2131-118-20041027) with ESMTP
          id <20100629154237.OZBY15102.schemailmta04.ci.com@schemailedgegx04.ci.com>
          for <u@ci.com>; Tue, 29 Jun 2010 10:42:37 -0500
Received: from blah.com ([1.1.1.1])
          by schemailedgegx04.ci.com
          (InterMail vG.1.02.00.04 201-2136-104-104-20050323) with ESMTP
          id <20100629154225.WEFB17009.schemailedgegx04.ci.com@blah.com>
          for <u@ci.com>; Tue, 29 Jun 2010 10:42:25 -0500
Received: from blah.com (snooki [10.12.126.68])
    by blah.com (Postfix) with ESMTP id 4BDAE19F546
    for <u@ci.com>; Tue, 29 Jun 2010 15:42:25 +0000 (UTC)
DKIM-Signature: v=1; a=rsa-sha256; c=simple/simple; d=blah.com;
    s=2010; t=1277826145; bh=wC3hHAhQgApcTmwQsi2F4OJf40rbyIek/WwIuzSc3V
    M=; h=Date:From:Reply-To:To:Message-ID:Subject:Mime-Version:
     Content-Type:Content-Transfer-Encoding:List-Unsubscribe; b=aw+Bhd8
    t1goZUXWBAHSrHaM1IdqhkXqF5WVMwGRYcnya4FHNw05XfpB3TTpTFda13DfhtziFRk
    zHSfiNbMapv7Vz+D3A/9NHg5nKahSMosZVTa0BfajYWNd1aY8JUWUlxdQHxQQ4ygCBj
    /MndJohtSm6K3gsqdIv88DNXdBGBEw=
Date: Tue, 29 Jun 2010 15:42:25 +0000
From: HomeRun <notification@blah.com>
Reply-To: HomeRun <notification+45b0d380@blah.com>
To: u@ci.com
Message-ID: <4c2a146147ac8_61ff157c4ec1652df@s.h.c.mail>
Subject: Your Friend F M wants you to join HomeRun
Mime-Version: 1.0
Content-Type: multipart/alternative;
 boundary="--==_mimepart_4c2a146141756_61ff157c4ec1649a8";
 charset=UTF-8
Content-Transfer-Encoding: 7bit
List-Unsubscribe: <mailto:unsubscribe+45b0d380@blah.com>



----==_mimepart_4c2a146141756_61ff157c4ec1649a8
Date: Tue, 29 Jun 2010 15:42:25 +0000
Mime-Version: 1.0
Content-Type: text/plain;
 charset=UTF-8
Content-Transfer-Encoding: base64
Content-ID: <4c2a146145451_61ff157c4ec165040@s.h.c.mail>

SGV5IGNpbmd1bGFybWVmYXJpZGEsCgpGYXJpZGEgTWFsaWsgdGhpbmtzIHlv
dSBzaG91bGQgYXBwbHkgdG8gam9pbiBIb21lUnVuLCB5b3VyIHBsYWNlIGZv
dC4sIFNhbiBGcmFuY2lzY28sIENBLCA5NDEyMywgVVNB


----==_mimepart_4c2a146141756_61ff157c4ec1649a8
Date: Tue, 29 Jun 2010 15:42:25 +0000
Mime-Version: 1.0
Content-Type: text/html;
 charset=UTF-8
Content-Transfer-Encoding: base64
Content-ID: <4c2a1461468ae_61ff157c4ec165194@s.h.c.mail>

PCFET0NUWVBFIGh0bWw+CjxodG1sPgo8aGVhZD4KPHRpdGxlPkhvbWVSdW4g
LSBZb3VyIEZyaWVuZCBGYXJpZGEgTWFsaWsgd2FudHMgeW91IHRvIGpvaW4g
cnVuLmNvbS9vLjQ1YjBkMzgwLmdpZicgd2lkdGg9JzEnIC8+CjwvdGQ+Cjwv
dHI+CjwvdGFibGU+CjwvdGQ+CjwvdHI+CjwvdGFibGU+CjwvZGl2Pgo8L2Jv
ZHk+CjwvaHRtbD4K


----==_mimepart_4c2a146141756_61ff157c4ec1649a8--

--===========================_ _= 6078796(15102)1277826164--
Comments
`,
  "Mail System Error - Returned Mail", "notification+promo@blah.com", "", "Postmaster@ci.com", "Mail Administrator", `This Message was undeliverable due to the following reason:


<u@ci.com> has restricted SMS e-mail

Please reply to <Postmaster@ci.com>
if you feel this message to be in error.
Reporting-MTA: dns; schemailmta04.ci.com
Arrival-Date: Tue, 29 Jun 2010 10:42:37 -0500
Received-From-MTA: dns; schemailedgegx04.ci.com (172.16.130.170)

Original-Recipient: rfc822;u@ci.com
Final-Recipient: RFC822; <u@ci.com>
Action: failed
Status: 5.3.0
`, "", []goodMailAttachments{}},

}
// bad mails

type badMailTypeTest struct {
  RawBody     string
}

var badMailTypeTests = []badMailTypeTest{
  {""},
  {"Invalid email body"},
}

// TESTS

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type ParserSuite struct{}
var _ = Suite(&ParserSuite{})

func (s *ParserSuite) SetUpTest(c *C) {
  // logger
  log.SetTarget(stdlog.New(os.Stdout, "", stdlog.LstdFlags))
  //log.SetTarget(stdlog.New(os.NewFile(uintptr(syscall.Stdout), os.DevNull), "", stdlog.LstdFlags))
  // uncomment for debug
  //log.Debug = true
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

// good emails

func (s *ParserSuite) TestGoodMailParser(c *C) {
  for _, mail := range goodMailTypeTests {
    testBody := strings.Replace(mail.RawBody, "\n", "\r\n", -1)
    // parse email
    envelop := &smtpd.BasicEnvelope{ MailboxID: 0, MailBody: []byte(testBody)}
    email, err := ParseMail(envelop)
    c.Assert(err, IsNil)
    if email == nil || err != nil {
      c.Errorf("Error in parsing email: %v", err)
    } else {
      c.Check(email.Subject, Equals, mail.Subject)
      c.Check(email.To.Address, Equals, mail.To)
      c.Check(email.To.Name, Equals, mail.ToName)
      c.Check(email.From.Address, Equals, mail.From)
      c.Check(email.From.Name, Equals, mail.FromName)
      if mail.Text == email.TextPart {
        c.Check(email.TextPart, Equals, mail.Text)
      } else {
        c.Check(email.TextPart, Equals, strings.Replace(mail.Text, "\n", "\r\n", -1))
      }

      c.Check(email.HtmlPart, Equals, strings.Replace(mail.Html, "\n", "\r\n", -1))
      if len(mail.Attachments) != len(email.Attachments) {
        c.Errorf("Unexpected value for Count of attachments; got %d but expected: %d, subject: %s",
          len(mail.Attachments), len(email.Attachments), email.Subject)
      }
      if len(mail.Attachments) > 0 {
        for i, attachment := range email.Attachments {
          c.Check(attachment.AttachmentFileName, Equals, mail.Attachments[i].Filename)
        }
      }
    }
  }
}

// parser bench

func (s *ParserSuite) BenchmarkParser(c *C) {
  for i := 0; i < c.N; i++ {
    mail := goodMailTypeTests[rand.Intn(len(goodMailTypeTests))]
    testBody := strings.Replace(mail.RawBody, "\n", "\r\n", -1)
    // parse email
    envelop := &smtpd.BasicEnvelope{ MailboxID: 0, MailBody: []byte(testBody)}
    _, err := ParseMail(envelop)
    if err != nil {
      c.Errorf("Error in parsing email: %v", err)
    }
  }
}

// bad emails

func (s *ParserSuite) TestBadMailParser(c *C) {
  for _, mail := range badMailTypeTests {
    testBody := strings.Replace(mail.RawBody, "\n", "\r\n", -1)
    // parse email
    envelop := &smtpd.BasicEnvelope{ MailboxID: 0, MailBody: []byte(testBody)}
    email, err := ParseMail(envelop)
    c.Assert(err, NotNil)
    if err == nil {
      c.Errorf("No error in parsing email: %v", email)
    }
  }
}
