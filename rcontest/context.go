package rcontest

import (
	"net"

	"github.com/gorcon/rcon"
)

// Context represents the context of the current RCON request. It holds request
// and conn objects and registered handler.
type Context struct {
	server  *Server
	conn    net.Conn
	request *rcon.Packet
	handler HandlerFunc
}

// Server returns the Server instance.
func (c *Context) Server() *Server {
	return c.server
}

// Conn returns current RCON connection.
func (c *Context) Conn() net.Conn {
	return c.conn
}

// Request returns received *rcon.Packet.
func (c *Context) Request() *rcon.Packet {
	return c.request
}

// Handler returns the matched handler.
func (c *Context) Handler() HandlerFunc {
	return c.handler
}
