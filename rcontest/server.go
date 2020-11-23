// package rcontest contains RCON server for RCON client testing.
// WIP: rcontest is not finally implemented. DO NOT USE IN PRODUCTION!

package rcontest

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/gorcon/rcon"
)

// Server is an RCON server listening on a system-chosen port on the
// local loopback interface, for use in end-to-end RCON tests.
type Server struct {
	addr        string
	listener    net.Listener
	handler     Handler
	connections map[net.Conn]struct{}
	quit        chan bool
	wg          sync.WaitGroup
	mu          sync.Mutex
	closed      bool
}

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("rcontest: failed to listen on a port: %v", err))
	}

	return l
}

// NewServer returns a running RCON Server or nil if an error occurred.
// The caller should call Close when finished, to shut it down.
func NewServer(handler HandlerFunc) *Server {
	server := NewUnstartedServer(handler)

	server.Start()

	return server
}

// NewUnstartedServer returns a new Server but doesn't start it.
// After changing its configuration, the caller should call Start.
// The caller should call Close when finished, to shut it down.
func NewUnstartedServer(handler HandlerFunc) *Server {
	if handler == nil {
		handler = commandHandler
	}

	server := Server{
		listener: newLocalListener(),
		handler: Handler{
			auth:    authHandler,
			command: handler,
		},
		connections: make(map[net.Conn]struct{}),
		quit:        make(chan bool),
	}

	return &server
}

// Start starts a server from NewUnstartedServer.
func (s *Server) Start() {
	if s.addr != "" {
		panic("server already started")
	}

	s.addr = s.listener.Addr().String()

	s.goServe()
}

// Close shuts down the MockServer.
func (s *Server) Close() {
	if s.closed {
		return
	}

	s.closed = true

	close(s.quit)

	s.listener.Close()

	// Waiting for server connections.
	s.wg.Wait()

	s.mu.Lock()
	for c := range s.connections {
		// Force-close any connections.
		s.closeConn(c)
	}
	s.mu.Unlock()
}

// Addr returns IPv4 string MockServer address.
func (s *Server) Addr() string {
	return s.addr
}

// serve handles incoming requests until a stop signal is given with Close.
func (s *Server) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isRunning() {
				panic(fmt.Errorf("serve error: %w", err))
			}

			return
		}

		s.wg.Add(1)

		go s.handle(conn)
	}
}

func (s *Server) goServe() {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		s.serve()
	}()
}

// handle handles incoming client conn.
func (s *Server) handle(conn net.Conn) {
	s.mu.Lock()
	s.connections[conn] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.closeConn(conn)
		s.wg.Done()
	}()

	for {
		request := &rcon.Packet{}
		if _, err := request.ReadFrom(conn); err != nil {
			if err == io.EOF {
				return
			}

			panic(fmt.Errorf("failed read request: %w", err))
		}

		switch request.Type {
		case rcon.SERVERDATA_AUTH:
			s.handler.auth(s, conn, request)
		case rcon.SERVERDATA_EXECCOMMAND:
			s.handler.command(s, conn, request)
		}
	}
}

// isRunning returns true if MockServer is running and false if is not.
func (s *Server) isRunning() bool {
	select {
	case <-s.quit:
		return false
	default:
		return true
	}
}

// closeConn closes a client conn and removes it from connections map.
func (s *Server) closeConn(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := conn.Close(); err != nil {
		panic(fmt.Errorf("close conn error: %w", err))
	}

	delete(s.connections, conn)
}
