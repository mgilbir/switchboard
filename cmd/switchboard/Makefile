install:
	go build
	mv switchboard /usr/local/bin/
	sudo cp conf/net.franchu.switchboard.plist /Library/LaunchDaemons/
	sudo launchctl unload /Library/LaunchDaemons/net.franchu.switchboard.plist
	sudo launchctl load /Library/LaunchDaemons/net.franchu.switchboard.plist
	sudo pkill switchboard

.PHONY: install
