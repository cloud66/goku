#!/usr/bin/ruby

Signal.trap "QUIT" do
	puts "received QUIT"
	puts "exiting now"
	abort
end

Signal.trap "TERM" do
	puts "received TERM"
	puts "exiting now"
	abort
end

Signal.trap "USR1" do
	puts "received USR1"
	puts "draining"
	$0='master process [draining]'
end

puts "Started"
$0='master process [active]'
while true do
	sleep(1)
end
