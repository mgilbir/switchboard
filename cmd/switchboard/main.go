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
)

var (
	defaults = map[string]interface{}{
		// Use Google public DNS as default if none is provided
		keyDefaultNameServers: []string{"8.8.8.8", "8.8.4.4"},
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

	s := switchboard.New(":12345")
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
			hSink := switchboard.NewSinkholeHandler(sinkhole)
			s.AddHandler(hSink)
		}
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
