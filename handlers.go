package switchboard

import (
	"fmt"
	"net"
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

type MappingHandler struct {
	path string
	ip   net.IP
}

func NewMappingHandler(path string, ip string) MappingHandler {
	ip = strings.TrimSpace(ip)

	return MappingHandler{
		path: path,
		ip:   net.ParseIP(ip),
	}
}

func (h MappingHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	//TODO: Do we want to do something different based on record type?
	m := &dns.Msg{}
	m.RecursionAvailable = true
	m.SetReply(r)

	header := dns.RR_Header{
		Name:   r.Question[0].Name,
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    5,
	}

	a := &dns.A{
		Hdr: header,
		A:   h.ip,
	}

	m.Answer = append(m.Answer, a)

	w.WriteMsg(m)
}

func (h MappingHandler) Path() string {
	return h.path
}
