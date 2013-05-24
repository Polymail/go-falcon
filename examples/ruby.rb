require 'net/smtp'

message = <<-END.split("\n").map!(&:strip).join("\n")
From: Private Person <me@fromdomain.com>
To: A Test User <test@todomain.com>
CC: <test2@todomain.com>
Subject: SMTP e-mail test

This is a test e-mail message.
END

Net::SMTP.start('localhost',
                2525,
                'localhost',
                'leo@leo.com', 'pass', :plain) do |smtp|
    smtp.send_message message, 'me@fromdomain.com',
                              ['test@todomain.com', 'test2@todomain.com']
end

Net::SMTP.start('localhost',
                2525,
                'localhost',
                'leo@leo.com', 'pass', :login) do |smtp|
    smtp.send_message message, 'me@fromdomain.com',
                              ['test@todomain.com', 'test2@todomain.com']
end

Net::SMTP.start('localhost',
                2525,
                'localhost',
                'leo@leo.com', 'pass', :cram_md5) do |smtp|
    smtp.send_message message, 'me@fromdomain.com',
                              ['test@todomain.com', 'test2@todomain.com']
end