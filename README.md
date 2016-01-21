# Go-Falcon [![Build Status](https://travis-ci.org/le0pard/go-falcon.png?branch=master)](https://travis-ci.org/le0pard/go-falcon)

SMTP server with POP3 and nginx proxy support, which store of mail messages in a relational database (PostgreSQL). Have support hooks with redis and http.

## Install

    make all

### Tests

    make test

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
  email_username character varying,
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

CREATE UNIQUE INDEX index_inboxes_on_email_username ON inboxes USING btree (email_username);

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

Test telnet (smtp):

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