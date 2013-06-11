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

Net::SMTP.start('localhost',
                1025,
                'localhost',
                username, password, :plain) do |smtp|
    smtp.send_message message, 'me@fromdomain.com',
                              ['test@todomain.com', 'test2@todomain.com', 'test3@todomain.com']
end

Net::SMTP.start('localhost',
                1025,
                'localhost',
                username, password, :login) do |smtp|
    smtp.send_message message, 'me@fromdomain.com',
                              ['test@todomain.com', 'test2@todomain.com', 'test3@todomain.com']
end


Net::SMTP.start('localhost',
                1025,
                'localhost',
                username, password, :cram_md5) do |smtp|
    smtp.send_message message, 'me@fromdomain.com',
                              ['test@todomain.com', 'test2@todomain.com', 'test3@todomain.com']
end
