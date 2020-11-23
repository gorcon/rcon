package rcontest

import (
	"bytes"
	"encoding/binary"
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

func authHandler(s *Server, conn io.Writer, request *rcon.Packet) {
	responseType := rcon.SERVERDATA_RESPONSE_VALUE
	responseID := request.ID
	responseBody := ""

	if request.Body() != "password" {
		if request.Body() == "timeout" {
			time.Sleep(rcon.DefaultDialTimeout + 1*time.Second)
		}

		// If authentication was failed, the ID must be assigned to -1.
		responseID = -1
		responseBody = string([]byte{0x00})
	} else {
		_, _ = rcon.NewPacket(responseType, responseID, responseBody).WriteTo(conn)
	}

	// Auth success.
	_, _ = rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, responseID, responseBody).WriteTo(conn)
}

func commandHandler(s *Server, conn io.Writer, request *rcon.Packet) {
	writeWithInvalidPadding := func(conn io.Writer, packet *rcon.Packet) error {
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

		return nil
	}

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
		if _, err := rcon.NewPacket(4, responseID, responseBody).WriteTo(conn); err != nil {
			return
		}

		responseBody = request.Body()
	case "padding":
		_ = writeWithInvalidPadding(conn, rcon.NewPacket(responseType, responseID, ""))

		return
	default:
		responseBody = "unknown command"
	}

	_, _ = rcon.NewPacket(responseType, responseID, responseBody).WriteTo(conn)
}
