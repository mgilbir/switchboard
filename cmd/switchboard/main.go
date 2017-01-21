package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/mgilbir/switchboard"
	"github.com/mgilbir/viper"
)

const (
	keyDefaultNameServers = "DefaultNameServers"
	keyProxy              = "proxy"
	keyBlacklists         = "blacklist"
	keyBindAddr           = "bind"
	keyMapping            = "mapping"
	keyApiServerAddr      = "apiBind"
)

var (
	defaults = map[string]interface{}{
		// Use Google public DNS as default if none is provided
		keyDefaultNameServers: []string{"8.8.8.8", "8.8.4.4"},
		keyBindAddr:           ":53",
		keyApiServerAddr:      ":8053",
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

	analytics := switchboard.NewAnalytics()

	s := switchboard.New(viper.GetString(keyBindAddr))
	errCh := make(chan error)

	go func(errCh chan error) {
		errCh <- s.ListenAndServe()
	}(errCh)

	// No handlers available yet.
	// This ensures that no DNS requests go through until a safe system is setup

	// Sinkhole handlers
	var blacklists []blacklistConfig
	err = viper.UnmarshalKey(keyBlacklists, &blacklists)
	if err != nil {
		log.Fatal(err)
	}

	for _, blacklist := range blacklists {
		// Proxy blacklist source domain
		u, err := url.Parse(blacklist.Src)
		if err != nil {
			log.Fatalf("Wrong configuration. Blacklist URL: %s is invalid. %s", blacklist.Src, err)
		}
		// Host may contain the port, strip it if present
		domain := strings.Split(u.Host, ":")[0]
		hProxy := switchboard.NewProxyHandler(domain, viper.GetStringSlice(keyDefaultNameServers)).WithAnalytics(analytics)
		s.AddHandler(hProxy)

		// Retrieve blacklist
		bl, err := switchboard.RetrieveBlacklist(blacklist.Src, blacklist.Category)
		if err != nil {
			log.Printf("Error retrieving blacklist: %s. %s. Proceeding without it\n", blacklist.Src, err)
		}

		for _, sinkhole := range bl.Domains() {
			hSink := switchboard.NewSinkholeHandler(sinkhole, bl.Category()).WithAnalytics(analytics)
			s.AddHandler(hSink)
		}
	}

	// Prepare and add handlers
	defaultHandler := switchboard.NewDefaultHandler(viper.GetStringSlice(keyDefaultNameServers))
	defaultHandler = defaultHandler.WithAnalytics(analytics)
	s.AddHandler(defaultHandler)

	// Proxy handlers
	var proxies []proxyConfig
	err = viper.UnmarshalKey(keyProxy, &proxies)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range proxies {
		hProxy := switchboard.NewProxyHandler(p.Domain, p.NameServers).WithAnalytics(analytics)
		s.AddHandler(hProxy)
	}

	// Mapping handlers
	for domain, ip := range viper.GetStringMapString(keyMapping) {
		hMap := switchboard.NewMappingHandler(domain, ip).WithAnalytics(analytics)
		s.AddHandler(hMap)
	}

	// Start API server
	api := switchboard.NewApi(analytics)
	apiServer := http.Server{
		Addr:    viper.GetString(keyApiServerAddr),
		Handler: api,
	}
	//TODO: handle errors
	go apiServer.ListenAndServe()

	err = <-errCh
	if err != nil {
		log.Fatal(err)
	}
}
