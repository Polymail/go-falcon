require 'net/smtp'

raise "args shoud be email ans pass" if ARGV.length < 2
username, password = ARGV[0], ARGV[1]

message = <<-END.split("\n").map!(&:strip).join("\n")
From: Private Person <me@fromdomain.com>
To: A Test User <test@todomain.com>
CC: <test2@todomain.com>
CC: <test3@todomain.com>
Subject: SMTP e-mail test

This is a test e-mail message.
END
=begin
message = <<-END.split("\n").map!(&:strip).join("\n")
From: Private Person <me@fromdomain.com>
To: A Test User <test@todomain.com>
Subject: Virus message

This is virus
$CEliacmaTrESTuScikgsn$FREE-TEST-SIGNATURE$EEEEE$
END
=end
Net::SMTP.start('falcon.rw.rw',
                2525,
                'falcon.rw.rw',
                username, password, :plain) do |smtp|
    smtp.send_message message, 'me@fromdomain.com',
                              ['test@todomain.com', 'test2@todomain.com', 'test3@todomain.com']
end

Net::SMTP.start('falcon.rw.rw',
                2525,
                'falcon.rw.rw',
                username, password, :login) do |smtp|
    smtp.send_message message, 'me@fromdomain.com',
                              ['test@todomain.com', 'test2@todomain.com', 'test3@todomain.com']
end

Net::SMTP.start('falcon.rw.rw',
                2525,
                'falcon.rw.rw',
                username, password, :cram_md5) do |smtp|
    smtp.send_message message, 'me@fromdomain.com',
                              ['test@todomain.com', 'test2@todomain.com', 'test3@todomain.com']
end
