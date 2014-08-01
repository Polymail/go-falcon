require 'net/pop'

raise "args shoud be email ans pass" if ARGV.length < 2
username, password = ARGV[0], ARGV[1]

Net::POP3.start('mailtrap.rw.rw', 110, username, password, true) do |pop|
  if pop.mails.empty?
    puts 'No mail.'
  else
    pop.each_mail do |m|   # or "pop.mails.each ..."   # (2)
      puts m.pop
      m.delete
    end
    puts "#{pop.mails.size} mails popped."
  end
end