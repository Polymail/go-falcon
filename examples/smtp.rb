require 'net/smtp'

raise "args shoud be email ans pass" if ARGV.length < 2
username, password = ARGV[0], ARGV[1]

message = <<-END.split("\n").map!(&:strip).join("\n")
From: Private Person <me@railsware.com>
To: A Test User <test@railsware.com>
To: <test2@railsware.com>
CC: <test3@todomain.com>
Subject: SMTP e-mail test

This is a test e-mail message.
END

# message = File.read('examples/test.eml')

=begin

message = <<-END.split("\n").map!(&:strip).join("\n")
Subject: Test spam mail (GTUBE)
Message-ID: <GTUBE1.1010101@example.net>
Date: Wed, 23 Jul 2003 23:30:00 +0200
From: Sender <sender@example.net>
To: Recipient <recipient@example.net>
Precedence: junk
MIME-Version: 1.0
Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: 7bit

This is the GTUBE, the
	Generic
	Test for
	Unsolicited
	Bulk
	Email

If your spam filter supports it, the GTUBE provides a test by which you
can verify that the filter is installed correctly and is detecting incoming
spam. You can send yourself a test mail containing the following string of
characters (in upper case and with no white spaces and line breaks):

XJS*C4JDBQADN1.NSBN3*2IDNEN*GTUBE-STANDARD-ANTI-UBE-TEST-EMAIL*C.34X

You should send this test mail from an account outside of your network.
END

message = <<-END.split("\n").map!(&:strip).join("\n")
From: Private Person <me@fromdomain.com>
To: A Test User <test@todomain.com>
Subject: Virus message

This is virus
$CEliacmaTrESTuScikgsn$FREE-TEST-SIGNATURE$EEEEE$
END
=end

=begin

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
    smtp.send_message message, "me#{rand(10)}@fromdomain.com",
                              ['test@todomain.com', 'test2@todomain.com', 'test3@todomain.com']
end
=end
#=begin
arr = []
1.times do
  arr << Thread.new do
    1.times do |i|
      Net::SMTP.start('falcon.rw.rw',
                      2525,
                      'falcon.rw.rw',
                      username, password, :cram_md5) do |smtp|
          smtp.send_message message, "me#{rand(10)}@fromdomain.com",
                                    ['test@todomain.com', 'test2@todomain.com', 'test3@todomain.com']
      end
      puts "#{i} sent"
    end
  end
end

arr.each {|t| t.join }
#=end
=begin

  Net::SMTP.start('localhost',
                  2525,
                  'localhost',
                  username, password, :plain) do |smtp|
      smtp.send_message message, "me#{rand(10)}@fromdomain.com",
                                ['test@todomain.com', 'test2@todomain.com', 'test3@todomain.com']
  end
  Net::SMTP.start('localhost',
                  2525,
                  'localhost',
                  username, password, :login) do |smtp|
      smtp.send_message message, "me#{rand(10)}@fromdomain.com",
                                ['test@todomain.com', 'test2@todomain.com', 'test3@todomain.com']
  end

  Net::SMTP.start('localhost',
                  2525,
                  'localhost',
                  username, password, :cram_md5) do |smtp|
      smtp.send_message message, "me#{rand(10)}@fromdomain.com",
                                ['test@todomain.com', 'test2@todomain.com', 'test3@todomain.com']
  end

=end
