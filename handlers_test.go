package switchboard

import (
	"net"
	"testing"

	"github.com/miekg/dns"
)

var (
	testHandlerRequestReceived = func(req, resp *dns.Msg) bool {
		return true
	}
)

func TestDummyHandlerAnalytics(t *testing.T) {
	probe := newTestAnalyticsProbe()
	probe.AddTest(analyticsTestHasDomain)
	probe.AddTest(analyticsTestHasNonEmptyQueryType)

	h := NewDummyHandler("test.com").WithAnalytics(probe)
	testQuery(h)

	if len(probe.m) != 1 {
		t.Errorf("Expected one entry, got %d", len(probe.m))
	}

	if !probe.Test() {
		t.Errorf("Expectations for analytics probe not fulfilled")
	}
}

func TestProxyHandlerAnalytics(t *testing.T) {
	probe := newTestAnalyticsProbe()
	probe.AddTest(analyticsTestHasDomain)
	probe.AddTest(analyticsTestHasNonEmptyQueryType)

	server := newDnsTestServer(t, testHandlerRequestReceived)
	defer server.Close()

	h := NewProxyHandler("test.com", []string{server.Addr}).WithAnalytics(probe)
	testQuery(h)

	if !server.handlerMux.TestOk {
		t.Errorf("Handler not OK")
	}

	if len(probe.m) != 1 {
		t.Errorf("Expected one entry, got %d", len(probe.m))
	}

	if !probe.Test() {
		t.Errorf("Expectations for analytics probe not fulfilled")
	}
}

func TestSinkholeHandlerAnalytics(t *testing.T) {
	probe := newTestAnalyticsProbe()
	probe.AddTest(analyticsTestHasDomain)
	probe.AddTest(analyticsTestHasNonEmptyQueryType)

	h := NewSinkholeHandler("test.com", "testCategory").WithAnalytics(probe)
	testQuery(h)

	if len(probe.m) != 1 {
		t.Errorf("Expected one entry, got %d", len(probe.m))
	}

	if !probe.Test() {
		t.Errorf("Expectations for analytics probe not fulfilled")
	}
}

func TestMappingHandlerAnalytics(t *testing.T) {
	probe := newTestAnalyticsProbe()
	probe.AddTest(analyticsTestHasDomain)
	probe.AddTest(analyticsTestHasNonEmptyQueryType)

	h := NewMappingHandler("test.com", "127.0.0.1").WithAnalytics(probe)
	testQuery(h)

	if len(probe.m) != 1 {
		t.Errorf("Expected one entry, got %d", len(probe.m))
	}

	if !probe.Test() {
		t.Errorf("Expectations for analytics probe not fulfilled")
	}
}

type testHandler struct {
	TestOk bool
	test   func(req, resp *dns.Msg) bool
}

func (th *testHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
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
		A:   net.ParseIP("1.2.3.4"),
	}

	m.Answer = append(m.Answer, a)

	w.WriteMsg(m)

	th.TestOk = th.test(r, m)
}

type dnsTestServer struct {
	Addr       string
	s          *dns.Server
	handlerMux *testHandler
}

func newDnsTestServer(t *testing.T, handler func(req, resp *dns.Msg) bool) *dnsTestServer {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	l, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		addr, err = net.ResolveUDPAddr("udp6", "[::1]:0")
		if err != nil {
			t.Fatal(err)
		}

		if l, err = net.ListenUDP("udp6", addr); err != nil {
			t.Fatalf("failed to listen on a port: %v", err)
		}
	}

	h := &testHandler{
		test: handler,
	}

	s := &dns.Server{
		PacketConn: l,
		Net:        "udp",
		UDPSize:    65535,
		Handler:    h,
	}

	go func() {
		if err := s.ActivateAndServe(); err != nil {
			t.Fatal(err)
		}
	}()

	testServer := &dnsTestServer{
		Addr:       l.LocalAddr().String(),
		s:          s,
		handlerMux: h,
	}

	return testServer
}

func (s dnsTestServer) Close() error {
	return s.s.Shutdown()
}

// Implements github.com/miekg/dns.ResponseWriter
type responseWriter struct {
}

// LocalAddr returns the net.Addr of the server
func (r *responseWriter) LocalAddr() net.Addr { return nil }

// RemoteAddr returns the net.Addr of the client that sent the current request.
func (r *responseWriter) RemoteAddr() net.Addr { return nil }

// WriteMsg writes a reply back to the client.
func (r *responseWriter) WriteMsg(*dns.Msg) error { return nil }

// Write writes a raw buffer back to the client.
func (r *responseWriter) Write([]byte) (int, error) { return 0, nil }

// Close closes the connection.
func (r *responseWriter) Close() error { return nil }

// TsigStatus returns the status of the Tsig.
func (r *responseWriter) TsigStatus() error { return nil }

// TsigTimersOnly sets the tsig timers only boolean.
func (r *responseWriter) TsigTimersOnly(bool) {}

// Hijack lets the caller take over the connection.
// After a call to Hijack(), the DNS package will not do anything with the connection.
func (r *responseWriter) Hijack() {}

func testQuery(handler Handler) {
	r := new(dns.Msg)
	r.SetQuestion("test.com.", dns.TypeA)

	w := new(responseWriter)
	handler.ServeDNS(w, r)
}

type testAnalyticsProbe struct {
	m     []AnalyticsMsg
	tests []func(AnalyticsMsg) bool
}

func newTestAnalyticsProbe() *testAnalyticsProbe {
	return &testAnalyticsProbe{}
}

func (p *testAnalyticsProbe) Handle(msg AnalyticsMsg) {
	p.m = append(p.m, msg)
}

func (p *testAnalyticsProbe) AddTest(f func(AnalyticsMsg) bool) {
	p.tests = append(p.tests, f)
}

func (p testAnalyticsProbe) Test() bool {
	for _, m := range p.m {
		for _, t := range p.tests {
			if !t(m) {
				return false
			}
		}
	}
	return true
}

func analyticsTestHasDomain(msg AnalyticsMsg) bool {
	return msg.Domain != ""
}

func analyticsTestHasNonEmptyQueryType(msg AnalyticsMsg) bool {
	return msg.QueryType != ""
}
