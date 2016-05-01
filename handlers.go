package switchboard

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

type Handler interface {
	Path() string
	ServeDNS(w dns.ResponseWriter, r *dns.Msg)
}

type DummyHandler struct {
	path string
}

func NewDummyHandler(path string) DummyHandler {
	return DummyHandler{path: path}
}

func (h DummyHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	fmt.Println(h.Path(), *r)
}

func (h DummyHandler) Path() string {
	return h.path
}

func NewDefaultHandler(nameservers []string) Handler {
	return NewProxyHandler(".", nameservers)
}

type ProxyHandler struct {
	path        string
	nameservers []string
}

func NewProxyHandler(path string, nameservers []string) ProxyHandler {
	ns := make([]string, len(nameservers))
	for i, v := range nameservers {
		if !strings.Contains(v, ":") {
			v = v + ":53"
		}
		ns[i] = v
	}
	return ProxyHandler{
		path:        path,
		nameservers: ns,
	}
}

func (h ProxyHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	c := &dns.Client{}
	response, _, err := c.Exchange(r, h.nameservers[0])
	if err == nil {
		if err := w.WriteMsg(response); err != nil {
			//TODO: improve error handling
			fmt.Println(err)
		}
	}
}

func (h ProxyHandler) Path() string {
	return h.path
}

type SinkholeHandler struct {
	path     string
	category string
}

func NewSinkholeHandler(path string, category string) SinkholeHandler {
	return SinkholeHandler{
		path:     path,
		category: category,
	}
}

func (h SinkholeHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	dns.HandleFailed(w, r)
}

func (h SinkholeHandler) Path() string {
	return h.path
}
