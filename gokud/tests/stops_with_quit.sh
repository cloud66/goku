#!/usr/bin/env ruby

Signal.trap("QUIT", proc { puts "Received QUIT"; abort })

while true do
	sleep(1)
end
