# Go-Falcon

storage of mail messages in a relational database


// https://github.com/bradfitz/go-smtpd/blob/master/smtpd/smtpd.go
// https://github.com/jda/go-lmtpd/blob/master/lmtpd/lmtpd.go

http://www.samlogic.net/articles/smtp-commands-reference-auth.htm

openssl s_client -starttls smtp -connect localhost:2525 -tls1 -crlf


```bash
telnet: > telnet localhost 2525
telnet: Trying 192.0.2.2...
telnet: Connected to mx1.example.com.
telnet: Escape character is '^]'.
server: 220 mx1.example.com ESMTP server ready Tue, 20 Jan 2004 22:33:36 +0200
client: HELO client.example.com
server: 250 mx1.example.com
client: MAIL from: <sender@example.com>
server: 250 Sender <sender@example.com> Ok
client: RCPT to: <recipient@example.com>
server: 250 Recipient <recipient@example.com> Ok
client: DATA
server: 354 Ok Send data ending with <CRLF>.<CRLF>
client: From: sender@example.com
client: To: recipient@example.com
client: Subject: Test message
client:
client: This is a test message.
client: .
server: 250 Message received: 20040120203404.CCCC18555.mx1.example.com@client.example.com
client: QUIT
server: 221 mx1.example.com ESMTP server closing connection
```