install:
	go build
	mv switchboard /usr/local/bin/
	sudo cp conf/net.franchu.switchboard.plist /Library/LaunchAgents/
	sudo launchctl load /Library/LaunchAgents/net.franchu.switchboard.plist
	sudo pkill switchboard

.PHONY: install
