package parser

import (
  stdlog "log"
  . "launchpad.net/gocheck"
  "os"
  "encoding/json"
  "strings"
  "testing"
  "math/rand"
  "path"
  "io/ioutil"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/protocol/smtpd"
)

// good mails

type goodMailAttachments struct {
  Filename      string
}

type goodMailTypeTest struct {
  Fixture     string

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
  {"1.eml", "SMTP e-mail test", "test@todomain.com", "A Test User", "me@fromdomain.com", "Private Person", "This is a test e-mail message.", "", []goodMailAttachments{}},

  {"2.eml", "test", "stephen.callaghan@greenfinch.ie", "", "mainstay@sherwoodcompliance.co.uk", "", "", "", []goodMailAttachments{
    {"OICLCostsPaymentProposal.csv"},
  }},

  {"3.eml", "illness", "Mr. X \"wrongquote@b.com\"", "", "sender@mail.com", "Mr. Sender", "  illness 26 Dec - 26 Dec 2007", "", []goodMailAttachments{}},

  {"4.eml", "illness notification ALPHÉE", "aaaa@bbbbbb.com", "", "sender@mail.com", "Mr. Sender", "illness 26 Dec - 26 Dec 2007", "", []goodMailAttachments{}},

  {"5.eml", "Welcome to Verical", "MichaelJWilliamstfb24d057-49fb-477d-8cf3-5357f2591641@test.com", "", "support@verical.com", "", "Please view the HTML version of this email.",
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

  {"6.eml", "Example subject line", "contactmichaelhart@gmail.com", "", "support@avocosecure.com", "support@avocosecure.com",
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

  {"7.eml", "Hello World", "", "", "", "", "Ã¿Ã´Ã¿Ã½", "", []goodMailAttachments{}},

  {"8.eml", "testing", "blah@example.com", "", "foo@example.com", "",
  "A fax has arrived from remote ID ''.\n------------------------------------------------------------\nTime: 3/9/2006 3:50:52 PM\nReceived from remote ID: \nInbound user ID XXXXXXXXXX, routing code XXXXXXXXX\nResult: (0/352;0/0) Successful Send\nPage record: 1 - 1\nElapsed time: 00:58 on channel 11\n",
  "", []goodMailAttachments{}},

  {"9.eml", "Re: Test: \"漢字\" mid \"漢字\" tail", "jamis@37signals.com", "", "jamis@37signals.com", "Jamis Buck", "대부분의 마찬가지로, 우리는 하나님을 믿습니다.\n\n제 이름은 Jamis입니다.", "", []goodMailAttachments{}},

  {"10.eml", "まみむめも", "raasdnil@gmail.com", "みける", "raasdnil@gmail.com", "Mikel Lindsaar",
"かきくえこ\n\n-- \nhttp://lindsaar.net/\nRails, RSpec and Life blog....\n",
  "", []goodMailAttachments{}},

  {"11.eml", "Eelanalüüsi päring", "jeff@37signals.com", "Jeffrey Hardy", "jeff@37signals.com", "Jeffrey Hardy", "", "", []goodMailAttachments{
    {"Eelanalüüsi päring.jpg"},
  }},

  {"12.eml", "this message JUST contains an attachment", "bob@domain.dom", "", "rfinnie@domain.dom", "Ryan Finnie", "", "", []goodMailAttachments{
    {"blah.gz"},
  }},

  {"13.eml", "testing", "blah@example.com", "", "foo@example.com", "", "This is the first part.\n", "", []goodMailAttachments{
    {"This is a test.txt"},
  }},

  {"14.eml",
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

  {"15.eml",
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

  {"16.eml", "Alerte suite a la recherche", "f.tete@immobilier-confiance.fr", "", "contact@immobilier-confiance.fr", "Immobilier Confiance", "",
  "Bonjour,\nSuite à la recherche ajoutée concernant le contact Test2 TEST\u003cbr/\u003eVoici les réultats : \u003cbr/\u003e\u003cbr/\u003eRésultats qui peuvent s'accorder aux termes de la recherche :\u003cbr/\u003e\u003ctable\u003e\u003ctr\u003e\u003cth\u003eRéférence\u003c/th\u003e\u003cth\u003eType de Bien\u003c/th\u003e\u003cth\u003ePrix Fai\u003c/th\u003e\u003cth\u003eNégociateur\u003c/th\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF901\u003c/td\u003e\u003ctd\u003eferme\u003c/td\u003e\u003ctd\u003e490000\u003c/td\u003e\u003ctd\u003eolivier Dal\u003c/td\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF905\u003c/td\u003e\u003ctd\u003emaison\u003c/td\u003e\u003ctd\u003e269000\u003c/td\u003e\u003ctd\u003efrédéric Ducrot\u003c/td\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF909\u003c/td\u003e\u003ctd\u003emaison\u003c/td\u003e\u003ctd\u003e234000\u003c/td\u003e\u003ctd\u003eolivier Dal\u003c/td\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF915\u003c/td\u003e\u003ctd\u003eloft\u003c/td\u003e\u003ctd\u003e115000\u003c/td\u003e\u003ctd\u003efrédéric Ducrot\u003c/td\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF9152\u003c/td\u003e\u003ctd\u003eloft\u003c/td\u003e\u003ctd\u003e125000\u003c/td\u003e\u003ctd\u003efrédéric Ducrot\u003c/td\u003e\u003c/tr\u003e\u003ctr\u003e\u003ctd\u003eREF927\u003c/td\u003e\u003ctd\u003emaison\u003c/td\u003e\u003ctd\u003e179000\u003c/td\u003e\u003ctd\u003eolivier Dal\u003c/td\u003e\u003c/tr\u003e\u003c/table\u003e",
  []goodMailAttachments{}},


  {"17.eml", "Testing outlook", "mikel@me.nowhere", "", "email_test@me.nowhere", "Mikel Lindsaar", "Hello\nThis is an outlook test\n\nSo there.\n\nMe.\n",
  "\u003c!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.0 Transitional//EN\"\u003e\n\u003cHTML\u003e\u003cHEAD\u003e\n\u003cMETA http-equiv=3DContent-Type content=3D\"text/html; =\ncharset=3Diso-8859-1\"\u003e\n\u003cMETA content=3D\"MSHTML 6.00.6000.16525\" name=3DGENERATOR\u003e\n\u003cSTYLE\u003e\u003c/STYLE\u003e\n\u003c/HEAD\u003e\n\u003cBODY bgColor=3D#ffffff\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003eHello\u003c/FONT\u003e\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003e\u003cSTRONG\u003eThis is an outlook=20\ntest\u003c/STRONG\u003e\u003c/FONT\u003e\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003e\u003cSTRONG\u003e\u003c/STRONG\u003e\u003c/FONT\u003e&nbsp;\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003e\u003cSTRONG\u003eSo there.\u003c/STRONG\u003e\u003c/FONT\u003e\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003e\u003c/FONT\u003e&nbsp;\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=3DArial size=3D2\u003eMe.\u003c/FONT\u003e\u003c/DIV\u003e\u003c/BODY\u003e\u003c/HTML\u003e\n\n",
  []goodMailAttachments{}},

  {"18.eml", "Re: TEST テスト テスト%F%9%H", "rudeboyjet@gmail.com", "", "atsushi@example.com", "Atsushi Yoshida", "Hello", "", []goodMailAttachments{}},

  {"19.eml", "Die Hasen und die Frösche (Microsoft Outlook 00)", "schmuergen@example.com", "", "doug@example.com", "Doug Sauder",
  "Die Hasen und die Frösche\n\nDie Hasen klagten einst über ihre mißliche Lage; \"wir leben\", sprach ein Redner, \"in steter Furcht vor Menschen und Tieren, eine Beute der Hunde, der Adler, ja fast aller Raubtiere! Unsere stete Angst ist ärger als der Tod selbst. Auf, laßt uns ein für allemal sterben.\" \n\nIn einem nahen Teich wollten sie sich nun ersäufen; sie eilten ihm zu; allein das außerordentliche Getöse und ihre wunderbare Gestalt erschreckte eine Menge Frösche, die am Ufer saßen, so sehr, daß sie aufs schnellste untertauchten. \n\n\"Halt\", rief nun eben dieser Sprecher, \"wir wollen das Ersäufen noch ein wenig aufschieben, denn auch uns fürchten, wie ihr seht, einige Tiere, welche also wohl noch unglücklicher sein müssen als wir.\" \n",
  "", []goodMailAttachments{}},

  {"20.eml", "Re: TEST テスト テスト%F%9%H", "rudeboyjet@gmail.com", "", "atsushi@example.com", "Atsushi Yoshida", "Hello", "", []goodMailAttachments{}},

  {"21.eml", "Test message from Microsoft Outlook 00", "jblow@example.com", "Joe Blow", "doug@example.com", "Doug Sauder",
  "\n\nThe Hare and the Tortoise \n \nA HARE one day ridiculed the short feet and slow pace of the Tortoise, who replied, laughing:  \"Though you be swift as the wind, I will beat you in a race.\"  The Hare, believing her assertion to be simply impossible, assented to the proposal; and they agreed that the Fox should choose the course and fix the goal.  On the day appointed for the race the two started together.  The Tortoise never for a moment stopped, but went on with a slow but steady pace straight to the end of the course.  The Hare, lying down by the wayside, fell fast asleep.  At last waking up, and moving as fast as he could, he saw the Tortoise had reached the goal, and was comfortably dozing after her fatigue.  \n \nSlow but steady wins the race.  \n\n\n",
  "\u003c!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.0 Transitional//EN\"\u003e\n\u003cHTML\u003e\u003cHEAD\u003e\n\u003cMETA content=\"text/html; charset=iso-8859-1\" http-equiv=Content-Type\u003e\n\u003cMETA content=\"MSHTML 5.00.2314.1000\" name=GENERATOR\u003e\u003c/HEAD\u003e\n\u003cBODY\u003e\n\u003cDIV\u003e\u003cFONT face=Arial size=2\u003e\u003cIMG align=baseline alt=\"blue ball\" border=0 \nhspace=0 src=\"cid:938014623@17052000-0f9b\"\u003e\u003c/FONT\u003e\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=Arial size=2\u003e\u003cBR\u003eThe Hare and the Tortoise \u003cBR\u003e&nbsp;\u003cBR\u003eA HARE \none day ridiculed the short feet and slow pace of the Tortoise, who replied, \nlaughing:&nbsp; \"Though you be swift as the wind, I will beat you in a \nrace.\"&nbsp; The Hare, believing her assertion to be simply impossible, assented \nto the proposal; and they agreed that the Fox should choose the course and fix \nthe goal.&nbsp; On the day appointed for the race the two started \ntogether.&nbsp; The Tortoise never for a moment stopped, but went on with a slow \nbut steady pace straight to the end of the course.&nbsp; The Hare, lying down by \nthe wayside, fell fast asleep.&nbsp; At last waking up, and moving as fast as he \ncould, he saw the Tortoise had reached the goal, and was comfortably dozing \nafter her fatigue.&nbsp; \u003cBR\u003e&nbsp;\u003cBR\u003eSlow but steady wins the race.&nbsp; \n\u003c/FONT\u003e\u003c/DIV\u003e\n\u003cDIV\u003e\u003cFONT face=Arial size=2\u003e\u003cBR\u003e&nbsp;\u003c/DIV\u003e\u003c/FONT\u003e\u003c/BODY\u003e\u003c/HTML\u003e\n",
  []goodMailAttachments{
    {"blueball.png"},
    {"greenball.png"},
    {"redball.png"},
  }},

  {"22.eml", "testing", "blah@example.com", "", "foo@example.com", "", `This is the first part.
Just attaching another PDF, here, to see what the message looks like,
and to see if I can figure out what is going wrong here.
`, "", []goodMailAttachments{
    {"broken.pdf"},
  }},

  {"23.eml",
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

  {"24.eml",
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

  {"25.eml", "Mail System Error - Returned Mail", "notification+promo@blah.com", "", "Postmaster@ci.com", "Mail Administrator",
"This Message was undeliverable due to the following reason:\r\n" +
"\r\n" +
"\r\n" +
"<u@ci.com> has restricted SMS e-mail\r\n" +
"\r\n" +
"Please reply to <Postmaster@ci.com>\r\n" +
"if you feel this message to be in error.\r\n" +
"Reporting-MTA: dns; schemailmta04.ci.com\r\n" +
"Arrival-Date: Tue, 29 Jun 2010 10:42:37 -0500\r\n" +
"Received-From-MTA: dns; schemailedgegx04.ci.com (172.16.130.170)\r\n" +
"\r\n" +
"Original-Recipient: rfc822;u@ci.com\r\n" +
"Final-Recipient: RFC822; <u@ci.com>\r\n" +
"Action: failed\r\n" +
"Status: 5.3.0\r\n" +
"Hey cingularmefarida,\n" +
"\n" +
"Farida Malik thinks you should apply to join HomeRun, your place fot., San Francisco, CA, 94123, USA",
"<!DOCTYPE html>\n" +
"<html>\n" +
"<head>\n" +
"<title>HomeRun - Your Friend Farida Malik wants you to join run.com/o.45b0d380.gif' width='1' />\n" +
"</td>\n" +
"</tr>\n" +
"</table>\n" +
"</td>\n" +
"</tr>\n" +
"</table>\n" +
"</div>\n" +
"</body>\n" +
"</html>\n", []goodMailAttachments{}},

{"26.eml", "Undelivered Mail Returned to Sender", "rahul.chaudhari@LL.com", "", "MAILER-DAEMON@lvmail01.LL.com", "",
"This is the mail system at host lvmail01.LL.com.\r\n" +
"\r\n" +
"I'm sorry to have to inform you that your message could not\r\n" +
"be delivered to one or more recipients. It's attached below.\r\n" +
"\r\n" +
"For further assistance, please send mail to postmaster.\r\n" +
"\r\n" +
"If you do so, please include this problem report. You can\r\n" +
"delete your own text from the attached returned message.\r\n" +
"\r\n" +
"                   The mail system\r\n" +
"\r\n" +
"<bbbbvhvbbvkjbhfbvbvjhb@gmail.com>: host\r\n" +
"    gmail-smtp-in.l.google.com[209.85.223.33] said: 550-5.1.1 The email account\r\n" +
"    that you tried to reach does not exist. Please try 550-5.1.1\r\n" +
"    double-checking the recipient's email address for typos or 550-5.1.1\r\n" +
"    unnecessary spaces. Learn more at                              550 5.1.1\r\n" +
"    http://mail.google.com/support/bin/answer.py?answer=6596 41si5422799iwn.27\r\n" +
"    (in reply to RCPT TO command)\r\n" +
"\r\n" +
"<bscdbcjhasbcjhbdscbhbsdhcbj@gmail.com>: host\r\n" +
"    gmail-smtp-in.l.google.com[209.85.223.33] said: 550-5.1.1 The email account\r\n" +
"    that you tried to reach does not exist. Please try 550-5.1.1\r\n" +
"    double-checking the recipient's email address for typos or 550-5.1.1\r\n" +
"    unnecessary spaces. Learn more at                              550 5.1.1\r\n" +
"    http://mail.google.com/support/bin/answer.py?answer=6596 41si5422799iwn.27\r\n" +
"    (in reply to RCPT TO command)\r\n" +
"\r\n" +
"<egyfefsdvsfvvhjsd@gmail.com>: host gmail-smtp-in.l.google.com[209.85.223.33]\r\n" +
"    said: 550-5.1.1 The email account that you tried to reach does not exist.\r\n" +
"    Please try 550-5.1.1 double-checking the recipient's email address for\r\n" +
"    typos or 550-5.1.1 unnecessary spaces. Learn more at\r\n" +
"    550 5.1.1 http://mail.google.com/support/bin/answer.py?answer=6596\r\n" +
"    41si5422799iwn.27 (in reply to RCPT TO command)\r\n" +
"\r\n" +
"<kfhejkfbsjkjsbhds@gmail.com>: host gmail-smtp-in.l.google.com[209.85.223.33]\r\n" +
"    said: 550-5.1.1 The email account that you tried to reach does not exist.\r\n" +
"    Please try 550-5.1.1 double-checking the recipient's email address for\r\n" +
"    typos or 550-5.1.1 unnecessary spaces. Learn more at\r\n" +
"    550 5.1.1 http://mail.google.com/support/bin/answer.py?answer=6596\r\n" +
"    41si5422799iwn.27 (in reply to RCPT TO command)\r\n" +
"\r\n" +
"<qfvhgsvhgsduiohncdhcvhsdfvsfygusd@gmail.com>: host\r\n" +
"    gmail-smtp-in.l.google.com[209.85.223.33] said: 550-5.1.1 The email account\r\n" +
"    that you tried to reach does not exist. Please try 550-5.1.1\r\n" +
"    double-checking the recipient's email address for typos or 550-5.1.1\r\n" +
"    unnecessary spaces. Learn more at                              550 5.1.1\r\n" +
"    http://mail.google.com/support/bin/answer.py?answer=6596 41si5422799iwn.27\r\n" +
"    (in reply to RCPT TO command)\r\n" +
"Reporting-MTA: dns; lvmail01.LL.com\r\n" +
"X-Postfix-Queue-ID: 9B7841BC027\r\n" +
"X-Postfix-Sender: rfc822; rahul.chaudhari@LL.com\r\n" +
"Arrival-Date: Tue, 23 Feb 2010 22:16:15 -0800 (PST)\r\n" +
"\r\n" +
"Final-Recipient: rfc822; bbbbvhvbbvkjbhfbvbvjhb@gmail.com\r\n" +
"Original-Recipient: rfc822;bbbbvhvbbvkjbhfbvbvjhb@gmail.com\r\n" +
"Action: failed\r\n" +
"Status: 5.1.1\r\n" +
"Remote-MTA: dns; gmail-smtp-in.l.google.com\r\n" +
"Diagnostic-Code: smtp; 550-5.1.1 The email account that you tried to reach does\r\n" +
"    not exist. Please try 550-5.1.1 double-checking the recipient's email\r\n" +
"    address for typos or 550-5.1.1 unnecessary spaces. Learn more at\r\n" +
"    550 5.1.1 http://mail.google.com/support/bin/answer.py?answer=6596\r\n" +
"    41si5422799iwn.27\r\n" +
"\r\n" +
"Final-Recipient: rfc822; bscdbcjhasbcjhbdscbhbsdhcbj@gmail.com\r\n" +
"Original-Recipient: rfc822;bscdbcjhasbcjhbdscbhbsdhcbj@gmail.com\r\n" +
"Action: failed\r\n" +
"Status: 5.1.1\r\n" +
"Remote-MTA: dns; gmail-smtp-in.l.google.com\r\n" +
"Diagnostic-Code: smtp; 550-5.1.1 The email account that you tried to reach does\r\n" +
"    not exist. Please try 550-5.1.1 double-checking the recipient's email\r\n" +
"    address for typos or 550-5.1.1 unnecessary spaces. Learn more at\r\n" +
"    550 5.1.1 http://mail.google.com/support/bin/answer.py?answer=6596\r\n" +
"    41si5422799iwn.27\r\n" +
"\r\n" +
"Final-Recipient: rfc822; egyfefsdvsfvvhjsd@gmail.com\r\n" +
"Original-Recipient: rfc822;egyfefsdvsfvvhjsd@gmail.com\r\n" +
"Action: failed\r\n" +
"Status: 5.1.1\r\n" +
"Remote-MTA: dns; gmail-smtp-in.l.google.com\r\n" +
"Diagnostic-Code: smtp; 550-5.1.1 The email account that you tried to reach does\r\n" +
"    not exist. Please try 550-5.1.1 double-checking the recipient's email\r\n" +
"    address for typos or 550-5.1.1 unnecessary spaces. Learn more at\r\n" +
"    550 5.1.1 http://mail.google.com/support/bin/answer.py?answer=6596\r\n" +
"    41si5422799iwn.27\r\n" +
"\r\n" +
"Final-Recipient: rfc822; kfhejkfbsjkjsbhds@gmail.com\r\n" +
"Original-Recipient: rfc822;kfhejkfbsjkjsbhds@gmail.com\r\n" +
"Action: failed\r\n" +
"Status: 5.1.1\r\n" +
"Remote-MTA: dns; gmail-smtp-in.l.google.com\r\n" +
"Diagnostic-Code: smtp; 550-5.1.1 The email account that you tried to reach does\r\n" +
"    not exist. Please try 550-5.1.1 double-checking the recipient's email\r\n" +
"    address for typos or 550-5.1.1 unnecessary spaces. Learn more at\r\n" +
"    550 5.1.1 http://mail.google.com/support/bin/answer.py?answer=6596\r\n" +
"    41si5422799iwn.27\r\n" +
"\r\n" +
"Final-Recipient: rfc822; qfvhgsvhgsduiohncdhcvhsdfvsfygusd@gmail.com\r\n" +
"Original-Recipient: rfc822;qfvhgsvhgsduiohncdhcvhsdfvsfygusd@gmail.com\r\n" +
"Action: failed\r\n" +
"Status: 5.1.1\r\n" +
"Remote-MTA: dns; gmail-smtp-in.l.google.com\r\n" +
"Diagnostic-Code: smtp; 550-5.1.1 The email account that you tried to reach does\r\n" +
"    not exist. Please try 550-5.1.1 double-checking the recipient's email\r\n" +
"    address for typos or 550-5.1.1 unnecessary spaces. Learn more at\r\n" +
"    550 5.1.1 http://mail.google.com/support/bin/answer.py?answer=6596\r\n" +
"    41si5422799iwn.27\r\n" +
"This is just testing.\r\n" +
"\r\n" +
"\r\n" +
"Thanks & Regards,\r\n" +
"Rahul P. Chaudhari\r\n" +
"Software Developer\r\n" +
"LIVIA India Private Limited\r\n" +
"\r\n" +
"Board Line - +91.22.6725 5100\r\n" +
"Hand Phone - +91.809 783 3437\r\n" +
"Web URL: www.LL.com \r\n", "", []goodMailAttachments{}},

{"27.eml", "Cron <root@blabla>", "root", "", "root (Cron Daemon)", "", "blabla-eeb74629", "", []goodMailAttachments{}},

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
    pathDir, _ := os.Getwd()
    RawBody, err := ioutil.ReadFile(path.Join(pathDir, "fixtures", mail.Fixture))
    if err != nil {
      panic("cannot read file " + mail.Fixture)
      return
    }
    testBody := strings.Replace(string(RawBody), "\n", "\r\n", -1)
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

      if mail.Html == email.HtmlPart {
        c.Check(email.HtmlPart, Equals, mail.Html)
      } else {
        c.Check(email.HtmlPart, Equals, strings.Replace(mail.Html, "\n", "\r\n", -1))
      }

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
  pathDir, _ := os.Getwd()

  for i := 0; i < c.N; i++ {
    mail := goodMailTypeTests[rand.Intn(len(goodMailTypeTests))]
    RawBody, err := ioutil.ReadFile(path.Join(pathDir, "fixtures", mail.Fixture))
    if err != nil {
      panic("cannot read file " + mail.Fixture)
      return
    }
    testBody := strings.Replace(string(RawBody), "\n", "\r\n", -1)
    // parse email
    envelop := &smtpd.BasicEnvelope{ MailboxID: 0, MailBody: []byte(testBody)}
    _, mailErr := ParseMail(envelop)
    if mailErr != nil {
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
