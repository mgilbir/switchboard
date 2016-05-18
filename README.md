# switchboard

A dynamic DNS server that proxies on a per domain basis to different nameservers
and allows to blacklist traffic.

# Example config

Put in /etc/switchboard/config.toml or ~/.config/switchboard/config.toml

```DefaultNameServers = ["8.8.8.8", "8.8.4.4"]
bind=":53"

[mapping]
"franchu.dev.local"="10.2.3.4"

[[blacklist]]
src="http://mirror1.malwaredomains.com/files/justdomains"
category="malware"

[[blacklist]]
src="https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist"
category="malware"

[[blacklist]]
src="https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt"
category="ads"

[[blacklist]]
src="http://hosts-file.net/ad_servers.txt"
category="ads"

[[proxy]]
# Cpy
domain="cpy.internal"
nameservers=["10.0.0.2"]

[[proxy]]
domain="compute.internal"
nameservers=[
		"10.0.0.2", # Cpy
	]
```

# TODO
 - Analytics UI
 - Persist analytics
 - API for config changes
 - Persist config
