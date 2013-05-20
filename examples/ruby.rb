require 'net/smtp'

message = <<-END.split("\n").map!(&:strip).join("\n")
From: Private Person <me@fromdomain.com>
To: A Test User <test@todomain.com>
Subject: SMTP e-mail test

This is a test e-mail message.
END

Net::SMTP.start('localhost',
                2525,
                'localhost',
                'username', 'password', :plain) do |smtp|
  smtp.send_message message, 'me@fromdomain.com',
                             'test@todomain.com'
end