# Go-Falcon [![Build Status](https://travis-ci.org/le0pard/go-falcon.png?branch=master)](https://travis-ci.org/le0pard/go-falcon)

storage of mail messages in a relational database

## Used libs

    go get launchpad.net/goyaml
    go get github.com/bmizerany/pq
    go get code.google.com/p/mahonia
    go get github.com/garyburd/redigo/redis

## Database SQL

```sql

CREATE TABLE inboxes
(
  id serial NOT NULL,
  company_id integer,
  name character varying(255),
  domain character varying(255) NOT NULL,
  username character varying(255) NOT NULL,
  password character varying(255) NOT NULL,
  max_size integer DEFAULT 0,
  created_at timestamp without time zone,
  updated_at timestamp without time zone,
  CONSTRAINT inboxes_pkey PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE inboxes
  OWNER TO leo;

-- Index: index_inboxes_on_company_id

-- DROP INDEX index_inboxes_on_company_id;

CREATE INDEX index_inboxes_on_company_id
  ON inboxes
  USING btree
  (company_id);

-- Index: index_inboxes_on_username

-- DROP INDEX index_inboxes_on_username;

CREATE UNIQUE INDEX index_inboxes_on_username
  ON inboxes
  USING btree
  (username COLLATE pg_catalog."default");

INSERT INTO mailboxes(domain, username, password, created_at, updated_at) VALUES ('leo.com', 'leo', 'pass', now(), now());



CREATE TABLE messages
(
  id serial NOT NULL,
  inbox_id integer,
  subject character varying(1000),
  sent_at timestamp without time zone,
  from_email character varying(255),
  from_name character varying(255),
  to_email character varying(255),
  to_name character varying(255),
  text_body text,
  html_body text,
  raw_body text,
  email_size integer DEFAULT 0,
  created_at timestamp without time zone,
  updated_at timestamp without time zone,
  CONSTRAINT messages_pkey PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE messages
  OWNER TO leo;

-- Index: index_messages_on_inbox_id

-- DROP INDEX index_messages_on_inbox_id;

CREATE INDEX index_messages_on_inbox_id
  ON messages
  USING btree
  (inbox_id);


CREATE TABLE messages_1 (CHECK ( inbox_id = 1 )) INHERITS (messages);

CREATE INDEX index_messages_1_on_inbox_id
  ON messages_1
  USING btree
  (inbox_id);



CREATE TABLE attachments
(
  id serial NOT NULL,
  inbox_id integer,
  message_id integer,
  filename character varying(1000),
  attachment_type character varying(255),
  content_type character varying(255),
  content_id character varying(255),
  transfer_encoding character varying(255),
  attachment_body bytea,
  attachment_size integer DEFAULT 0,
  created_at timestamp without time zone,
  updated_at timestamp without time zone,
  CONSTRAINT attachments_pkey PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE attachments
  OWNER TO leo;

-- Index: index_attachments_on_message_id

-- DROP INDEX index_attachments_on_message_id;

CREATE INDEX index_attachments_on_message_id
  ON attachments
  USING btree
  (message_id);




CREATE TABLE attachments_1 (CHECK ( inbox_id = 1 )) INHERITS (attachments);

CREATE INDEX index_attachments_1_on_inbox_id
  ON attachments_1
  USING btree
  (inbox_id);

CREATE INDEX index_attachments_1_on_content_id
  ON attachments_1
  USING btree
  (content_id);

CREATE INDEX index_attachments_1_on_attachment_type
  ON attachments_1
  USING btree
  (attachment_type);


```

## Test

    go test -v ./...

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

