package switchboard

import (
	"github.com/miekg/dns"
)

type Server struct {
	tcpServer  *dns.Server
	udpServer  *dns.Server
	handlerMux *dns.ServeMux
}

func New(addr string) *Server {
	handlerMux := dns.NewServeMux()
	tcpServer := &dns.Server{
		Addr:    addr,
		Net:     "tcp",
		Handler: handlerMux,
	}

	udpServer := &dns.Server{
		Addr:    addr,
		Net:     "udp",
		UDPSize: 65535,
		Handler: handlerMux,
	}

	return &Server{
		tcpServer:  tcpServer,
		udpServer:  udpServer,
		handlerMux: handlerMux,
	}
}

func (s *Server) AddHandler(handler Handler) error {
	s.handlerMux.Handle(handler.Path(), handler)
	return nil
}

func (s *Server) ListenAndServe() error {
	errCh := make(chan error)
	go func(errCh chan error) {
		errCh <- s.tcpServer.ListenAndServe()
	}(errCh)

	go func(errCh chan error) {
		errCh <- s.udpServer.ListenAndServe()
	}(errCh)

	for err := range errCh {
		return err
	}
	return nil
}
