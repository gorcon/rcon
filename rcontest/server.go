// package rcontest contains RCON server for RCON client testing.
// WIP: rcontest is not finally implemented. DO NOT USE IN PRODUCTION!

package rcontest

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/gorcon/rcon"
)

const MockPasswordRCON = "password"

// Server is a mock Source RCON Protocol server.
type Server struct {
	listener    net.Listener
	handler     Handler
	connections map[net.Conn]struct{}
	wg          sync.WaitGroup
	mu          sync.Mutex
	errors      chan error
	quit        chan bool
}

// NewServer returns a running RCON Server or nil if an error occurred.
func NewServer(handlers ...HandlerFunc) (*Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	if len(handlers) == 0 {
		handlers = []HandlerFunc{commandHandler}
	}

	server := &Server{
		listener: listener,
		handler: Handler{
			auth:    authHandler,
			command: handlers[0],
		},
		connections: make(map[net.Conn]struct{}),
		errors:      make(chan error, 10),
		quit:        make(chan bool),
	}

	server.wg.Add(1)

	go server.serve()

	return server, nil
}

// MustNewServer returns a running RCON Server or panic if an error occurred.
func MustNewServer() *Server {
	server, err := NewServer()
	if err != nil {
		panic(err)
	}

	return server
}

// Close shuts down the MockServer.
func (s *Server) Close() error {
	close(s.quit)

	err := s.listener.Close()

	// Waiting for server connections.
	s.wg.Wait()

	// And close remaining connections.
	s.mu.Lock()
	// TODO: remove extra `close connection error` message.
	for c := range s.connections {
		// Close connections and add original error if occurred.
		if err2 := c.Close(); err2 != nil {
			if err == nil {
				err = fmt.Errorf("close connection error: %w", err2)
			} else {
				err = fmt.Errorf("close connection error: %s. Previous error: %w", err2, err)
			}
		}
	}
	s.mu.Unlock()

	close(s.errors)

	return err
}

// MustClose shuts down the RCON Server or panic.
func (s *Server) MustClose() {
	if s == nil {
		return
	}

	if err := s.Close(); err != nil {
		panic(err)
	}

	for err := range s.errors {
		panic(err)
	}
}

// Addr returns IPv4 string MockServer address.
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

// serve handles incoming requests until a stop signal is given with Close.
func (s *Server) serve() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isRunning() {
				s.reportError(fmt.Errorf("serve error: %w", err))
			}

			return
		}

		s.wg.Add(1)

		go s.handle(conn)
	}
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

			s.reportError(fmt.Errorf("handle read request error: %w", err))

			return
		}

		switch request.Type {
		case rcon.SERVERDATA_AUTH:
			if err := s.handler.auth(s, conn, request); err != nil {
				s.reportError(fmt.Errorf("handle write response error: %w", err))

				return
			}
		case rcon.SERVERDATA_EXECCOMMAND:
			if err := s.handler.command(s, conn, request); err != nil {
				s.reportError(fmt.Errorf("handle write response error: %w", err))

				return
			}
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

// write writes packets to conn. Replaces packets ids to mirrored id from request.
func (s *Server) write(conn io.Writer, id int32, packets ...*rcon.Packet) error {
	for _, packet := range packets {
		packet.ID = id

		if _, err := packet.WriteTo(conn); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) writeWithInvalidPadding(conn io.Writer, id int32, packets ...*rcon.Packet) error {
	for _, packet := range packets {
		packet.ID = id

		buffer := bytes.NewBuffer(make([]byte, 0, packet.Size+4))

		if err := binary.Write(buffer, binary.LittleEndian, packet.Size); err != nil {
			return err
		}

		if err := binary.Write(buffer, binary.LittleEndian, packet.ID); err != nil {
			return err
		}

		if err := binary.Write(buffer, binary.LittleEndian, packet.Type); err != nil {
			return err
		}

		// Write command body, null terminated ASCII string and an empty ASCIIZ string.
		// Second padding byte is incorrect.
		if _, err := buffer.Write(append([]byte(packet.Body()), 0x00, 0x01)); err != nil {
			return err
		}

		if _, err := buffer.WriteTo(conn); err != nil {
			return err
		}
	}

	return nil
}

// closeConn closes a client conn and removes it from connections map.
func (s *Server) closeConn(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := conn.Close(); err != nil {
		s.reportError(fmt.Errorf("close conn error: %w", err))
	}

	delete(s.connections, conn)
}

// reportError writes error to errors channel.
func (s *Server) reportError(err error) {
	if err == nil {
		return
	}

	select {
	case s.errors <- err:
	default:
		panic("errors channel is locked")
	}
}
