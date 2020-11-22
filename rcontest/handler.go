package rcontest

import (
	"fmt"
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
type HandlerFunc func(s *Server, conn io.Writer, request *rcon.Packet) error

func authHandler(s *Server, conn io.Writer, request *rcon.Packet) error {
	responseType := rcon.SERVERDATA_RESPONSE_VALUE
	responseID := request.ID
	responseBody := ""

	if request.Body() != MockPasswordRCON {
		if request.Body() == "timeout" {
			time.Sleep(rcon.DefaultDialTimeout + 1*time.Second)
		}

		// If authentication was failed, the ID must be assigned to -1.
		responseID = -1
		responseBody = string([]byte{0x00})
	} else {
		_ = s.write(conn, responseID, rcon.NewPacket(responseType, responseID, responseBody))
	}

	// Auth success.
	return s.write(conn, responseID, rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, responseID, responseBody))
}

func commandHandler(s *Server, conn io.Writer, request *rcon.Packet) error {
	responseType := rcon.SERVERDATA_RESPONSE_VALUE
	responseID := request.ID
	responseBody := ""

	switch request.Body() {
	case "help":
		responseBody = "lorem ipsum dolor sit amet"
	case "deadline":
		time.Sleep(rcon.DefaultDeadline + 1*time.Second)

		responseBody = request.Body()
	case "rust":
		// Write specific Rust package.
		if err := s.write(conn, responseID, rcon.NewPacket(4, responseID, responseBody)); err != nil {
			return fmt.Errorf("handle write response error: %w", err)
		}

		responseBody = request.Body()
	case "padding":
		return s.writeWithInvalidPadding(conn, responseID, rcon.NewPacket(responseType, responseID, ""))
	default:
		responseBody = "unknown command"
	}

	return s.write(conn, responseID, rcon.NewPacket(responseType, responseID, responseBody))
}
