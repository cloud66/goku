#!/usr/bin/env ruby

Signal.trap("QUIT", proc { puts "Received QUIT" })
Signal.trap("TERM", proc { puts "Received TERM"; abort })

while true do
	sleep(1)
end
