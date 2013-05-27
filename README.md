# Go-Falcon

storage of mail messages in a relational database

## Libs

    go get launchpad.net/goyaml
    go get github.com/bmizerany/pq

## AUTH CRAM-MD5

http://www.samlogic.net/articles/smtp-commands-reference-auth.htm

## SQL

```sql

CREATE TABLE mailboxes
(
  id serial NOT NULL,
  email character varying(255) NOT NULL DEFAULT ''::character varying,
  raw_password character varying(255) NOT NULL DEFAULT ''::character varying,
  CONSTRAINT users_pkey PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE mailboxes
  OWNER TO leo;

CREATE UNIQUE INDEX index_mailboxes_on_email
  ON mailboxes
  USING btree
  (email COLLATE pg_catalog."default");

CREATE INDEX index_mailboxes_on_raw_password
  ON mailboxes
  USING btree
  (raw_password COLLATE pg_catalog."default");

INSERT INTO mailboxes(email, raw_password) VALUES ('leo@leo.com', 'secret');



```

## Test

Test telnet:

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

Test Tls:

```bash
openssl s_client -starttls smtp -connect localhost:2525 -tls1 -crlf
```


## Postfix


in /etc/postfix/master.cf add:

 dbmail-lmtp     unix    -       -       n       -       -       lmtp
If you want verbose output in the mail log, add -v to lmtp, like this:

 dbmail-lmtp     unix    -       -       n       -       -       lmtp -v
Note that this setting will result in a LOT of extra output in your logs.

If one or more destinations in your mydestination list are not DNS-resolvable, DNS lookups must be disabled for the dbmail-lmtp daemon, like this:

dbmail-lmtp     unix    -       -       n       -       -       lmtp
        -o disable_dns_lookups=yes
If you want to send all the email the MTA accepts to DBMail, use the following setting in /etc/postfix/main.cf:

virtual_transport = dbmail-lmtp:<host>:<port>
If you want to decide if DBMail should be used per domain please add the following in /etc/postfix/transport:

<domain>        dbmail-lmtp:<host>:<port>
Where <domain> should replaced by the mail domain you receive mail for. It is possible to have several domain entries here. For <host> and <port> fill out the host and port on which the dbmail-lmtp daemon runs. If unsure about which port they run on, check your dbmail.conf file. The standard port for the lmtp service is port 24. An example of a transport file is below:

example.com             dbmail-lmtp:localhost:24
anotherexample.com      dbmail-lmtp:localhost:24
Postfix needs to lookup if the recipient domains exist. Otherwise, Postfix will reject your DBMail recipients with a “User unknown in local recipient table” error.

To do this, you need to enable the MySQL- or PGSQL-module in Postfix and add

virtual_mailbox_domains = mysql:/etc/postfix/sql-virtual_mailbox_domains.cf
or for PostgreSQL

virtual_mailbox_domains = pgsql:/etc/postfix/sql-virtual_mailbox_domains.cf
in /etc/postfix/main.cf

After that create the file and add the following MySQL-statements (adjust it to your needs if you use Postgres):

user     = <SQL-username>
password = <SQL-password>
hosts    = <SQL-host>
dbname   = <SQL-database>
query    = SELECT DISTINCT 1 FROM dbmail_aliases WHERE SUBSTRING_INDEX(alias, '@', -1) = '%s';
For postgresql replace the query by

query    = SELECT DISTINCT 1  FROM dbmail_aliases WHERE SUBSTRING(alias FROM POSITION('@' in alias)+1) = '%s';
The query MUST return a non-domain value (just use “1”) if the queried domain exists or a empty set if it doesn't exist. Returning domains here will moving companies break the delivery chain of Posfix!

Note: In case you use mail addresses as usernames in DBMail, you also have to include 'dbmail_user' in the SQL-query.

If you also want to let Postfix check the recipient addresses to reduce load on DBMail by false email addresses, add

virtual_mailbox_maps = mysql:/etc/postfix/sql-virtual_mailbox_maps.cf
or for PostgreSQL

virtual_mailbox_maps = pgsql:/etc/postfix/sql-virtual_mailbox_maps.cf
in /etc/postfix/main.cf and add the following content (MySQL):

user     = <SQL-username>
password = <SQL-password>
hosts    = <SQL-host>
dbname   = <SQL-database>
query    = SELECT 1 FROM dbmail_aliases WHERE alias='%s';
For postgresql replace the query by

query    = SELECT DISTINCT 1  FROM dbmail_aliases WHERE alias= '%s';
Make sure not to return the addresses !!!

If you want to use wildcards (user@ and @domain) with postgresql, use this query:

query    = SELECT DISTINCT 1 FROM dbmail_aliases where alias='%s' OR alias=SUBSTRING('%s' FROM POSITION('@' in '%s')) OR ( ( SELECT DISTINCT 1 FROM dbmail_aliases WHERE SUBSTRING(alias FROM POSITION('@' in alias)+1) = SUBSTRING('%s' FROM POSITION('@' in '%s')+1) )=1 AND alias=SUBSTRING('%s' FOR POSITION('@' in '%s')));
It searches for the exact match (my@cow.com), then a matching domain wildcard (@cow.com). The subselect in the middle prevents the system from accepting mail for valid users at invalid domains (my@moose.com). The last match (AND) allows user wildcards for accepted domains.

The interaction here seems strange, but it should match the dbmail rules:

exact match
user wildcard with valid domain (domain mentioned somewhere in the alias table)
domain wildcard


in /etc/postfix/master.cf add:

 dbmail-smtp    unix  -       n       n       -       -       pipe
         flags=  user=<dbmailuser>:<dbmailgroup>
         argv=/usr/local/sbin/dbmail-smtp -d ${recipient} -r ${sender}
where <dbmailuser> and <dbmailgroup> should be the user and group the dbmail-smtp program should run as. The ${recipient} and ${sender} fields are filled in by Postfix.

To send all email to DBMail, add this in /etc/postfix/main.cf

 virtual_transport = dbmail-smtp:
If you want to decide whether or not to send to DBMail per domain, add this in /etc/postfix/transport:

 <domain>        dbmail-smtp:
See the section on running Postfix with LMTP if you don't understand the transport file.

now run:

 # postmap /etc/postfix/transport
 # postfix reload
And your mail will be delivered!

