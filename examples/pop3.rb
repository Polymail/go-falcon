require 'net/pop'

Net::POP3.start('falcon.rw.rw', 1110, '1ca5d2bea8f8c', 'a26d0a622ceefc', true) do |pop|
  if pop.mails.empty?
    puts 'No mail.'
  else
    i = 0
    pop.each_mail do |m|   # or "pop.mails.each ..."   # (2)
      puts m.pop
      m.delete
      i += 1
    end
    puts "#{pop.mails.size} mails popped."
  end
end