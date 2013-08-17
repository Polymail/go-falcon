require 'net/pop'

Net::POP3.start('falcon.rw.rw', 110, '90681fe5f57d5', '3e526ed313d8dc', true) do |pop|
  if pop.mails.empty?
    puts 'No mail.'
  else
    i = 0
    pop.each_mail do |m|   # or "pop.mails.each ..."   # (2)
      puts m.pop
      m.delete
      break if i > 1
      i += 1
    end
    puts "#{pop.mails.size} mails popped."
  end
end