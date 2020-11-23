package rcontest

import (
	"io"
	"time"

	"github.com/gorcon/rcon"
)

// A Handler responds to an RCON request.
type Handler struct {
	auth    HandlerFunc
	command HandlerFunc
}

// HandlerFunc defines a function to serve RCON requests.
type HandlerFunc func(s *Server, conn io.Writer, request *rcon.Packet)

func defaultAuthHandler(s *Server, conn io.Writer, request *rcon.Packet) {
	if s.settings.AuthResponseDelay != 0 {
		time.Sleep(s.settings.AuthResponseDelay)
	}

	if request.Body() == s.settings.Password {
		// First respond with empty SERVERDATA_RESPONSE_VALUE.
		_, _ = rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, request.ID, "").WriteTo(conn)

		// Respond with auth success.
		_, _ = rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, request.ID, "").WriteTo(conn)
	} else {
		// If authentication was failed, the ID must be assigned to -1.
		_, _ = rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, -1, string([]byte{0x00})).WriteTo(conn)
	}
}

func defaultCommandHandler(s *Server, conn io.Writer, request *rcon.Packet) {
	_, _ = rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, request.ID, "").WriteTo(conn)
}
