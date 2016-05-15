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
	path      string
	analytics AnalyticsHandlerFunc
}

func NewDummyHandler(path string) DummyHandler {
	return DummyHandler{path: path, analytics: NoOpAnalytics}
}

func (p DummyHandler) WithAnalytics(h AnalyticsHandler) DummyHandler {
	p.analytics = h.Handle
	return p
}

func (h DummyHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	qName, qType, err := extractQueryInfo(r)
	if err != nil {
		dns.HandleFailed(w, r)
		return
	}

	defer h.analytics(AnalyticsMsg{Domain: qName, QueryType: qType, Category: "DUMMY", Time: Now()})

	fmt.Println(h.Path(), *r)
}

func (h DummyHandler) Path() string {
	return h.path
}

func NewDefaultHandler(nameservers []string) ProxyHandler {
	h := NewProxyHandler(".", nameservers)
	return h
}

type ProxyHandler struct {
	path        string
	nameservers []string
	proxyType   string
	analytics   AnalyticsHandlerFunc
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
		analytics:   NoOpAnalytics,
	}
}

func (p ProxyHandler) WithAnalytics(h AnalyticsHandler) ProxyHandler {
	p.analytics = h.Handle
	return p
}

func (h ProxyHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	qName, qType, err := extractQueryInfo(r)
	if err != nil {
		dns.HandleFailed(w, r)
		return
	}

	defer h.analytics(AnalyticsMsg{Domain: qName, QueryType: qType, Time: Now()})
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
	path      string
	category  string
	analytics AnalyticsHandlerFunc
}

func NewSinkholeHandler(path string, category string) SinkholeHandler {
	return SinkholeHandler{
		path:      path,
		category:  category,
		analytics: NoOpAnalytics,
	}
}

func (p SinkholeHandler) WithAnalytics(h AnalyticsHandler) SinkholeHandler {
	p.analytics = h.Handle
	return p
}

func (h SinkholeHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	qName, qType, err := extractQueryInfo(r)
	if err != nil {
		dns.HandleFailed(w, r)
		return
	}

	defer h.analytics(AnalyticsMsg{Domain: qName, QueryType: qType, Category: h.category, Time: Now()})
	dns.HandleFailed(w, r)
}

func (h SinkholeHandler) Path() string {
	return h.path
}

type MappingHandler struct {
	path      string
	ip        net.IP
	analytics AnalyticsHandlerFunc
}

func NewMappingHandler(path string, ip string) MappingHandler {
	ip = strings.TrimSpace(ip)

	return MappingHandler{
		path:      path,
		ip:        net.ParseIP(ip),
		analytics: NoOpAnalytics,
	}
}

func (p MappingHandler) WithAnalytics(h AnalyticsHandler) MappingHandler {
	p.analytics = h.Handle
	return p
}

func (h MappingHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	qName, qType, err := extractQueryInfo(r)
	if err != nil {
		dns.HandleFailed(w, r)
		return
	}

	defer h.analytics(AnalyticsMsg{Domain: qName, QueryType: qType, Time: Now()})

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

func extractQueryInfo(m *dns.Msg) (string, string, error) {
	q := m.Question[0]
	qType, ok := dns.TypeToString[q.Qtype]
	if !ok {
		return q.Name, "", fmt.Errorf("Unknown query type: %d", q.Qtype)
	}
	return q.Name, qType, nil
}
