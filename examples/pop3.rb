require 'net/pop'

pop = Net::POP3.new('localhost', 1110)
pop.start('187950561efb4', '2fde0a47e988ef')             # (1)
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
pop.finish