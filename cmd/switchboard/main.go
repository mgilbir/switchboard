package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/mgilbir/switchboard"
	"github.com/spf13/viper"
)

const (
	keyDefaultNameServers = "DefaultNameServers"
	keyProxy              = "proxy"
	keyBlacklists         = "blacklist"
	keyBindAddr           = "bind"
	keyMapping            = "mapping"
)

var (
	defaults = map[string]interface{}{
		// Use Google public DNS as default if none is provided
		keyDefaultNameServers: []string{"8.8.8.8", "8.8.4.4"},
		keyBindAddr:           ":53",
	}
)

type proxyConfig struct {
	Domain      string
	NameServers []string
}

type blacklistConfig struct {
	Src      string
	Category string
}

func main() {
	for k, v := range defaults {
		viper.SetDefault(k, v)
	}

	// Read the configuration
	viper.SetConfigName("config")                     // name of config file (without extension)
	viper.AddConfigPath("/etc/switchboard/")          // path to look for the config file in
	viper.AddConfigPath("$HOME/.config/switchboard/") // call multiple times to add many search paths
	viper.AddConfigPath(".")                          // optionally look for config in the working directory
	err := viper.ReadInConfig()                       // Find and read the config file
	if err != nil {                                   // Handle errors reading the config file
		switch err.(type) {
		case viper.UnsupportedConfigError:
			var b bytes.Buffer
			for k, v := range defaults {
				b.WriteString(fmt.Sprintf("%s:%v\n", k, v))
			}
			fmt.Printf("No valid configuration file found.\nUsing defaults:\n%s", b.String())
		default:
			log.Fatal(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	}

	s := switchboard.New(viper.GetString(keyBindAddr))
	// Prepare and add handlers
	defaultHandler := switchboard.NewDefaultHandler(viper.GetStringSlice(keyDefaultNameServers))
	s.AddHandler(defaultHandler)

	// Proxy handlers
	var proxies []proxyConfig
	err = viper.UnmarshalKey(keyProxy, &proxies)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range proxies {
		hProxy := switchboard.NewProxyHandler(p.Domain, p.NameServers)
		s.AddHandler(hProxy)
	}

	// Sinkhole handlers
	var blacklists []blacklistConfig
	err = viper.UnmarshalKey(keyBlacklists, &blacklists)
	if err != nil {
		log.Fatal(err)
	}

	for _, blacklist := range blacklists {
		bl, err := switchboard.RetrieveBlacklist(blacklist.Src, blacklist.Category)
		if err != nil {
			log.Printf("Error retrieving blacklist: %s. %s. Proceeding without it\n", blacklist.Src, err)
		}

		for _, sinkhole := range bl.Domains() {
			hSink := switchboard.NewSinkholeHandler(sinkhole, bl.Category())
			s.AddHandler(hSink)
		}
	}

	// Mapping handlers
	for domain, ip := range viper.GetStringMapString(keyMapping) {
		hMap := switchboard.NewMappingHandler(domain, ip)
		s.AddHandler(hMap)
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
