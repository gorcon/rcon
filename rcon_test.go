package rcon

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDial(t *testing.T) {
	server, err := NewMockServer()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		assert.NoError(t, server.Close())
		close(server.errors)
		for err := range server.errors {
			assert.NoError(t, err)
		}
	}()

	// Test connection refused.
	func() {
		conn, err := Dial("127.0.0.2:12345", MockPassword)
		if !assert.Error(t, err) {
			// Close connection if established.
			assert.NoError(t, conn.Close())
		}
		assert.EqualError(t, err, "dial tcp 127.0.0.2:12345: connect: connection refused")
	}()

	// Test connection timeout.
	func() {
		conn, err := Dial(server.Addr(), "timeout", SetDialTimeout(5*time.Second))
		if !assert.Error(t, err) {
			// Close connection if established.
			assert.NoError(t, conn.Close())
		}
		assert.EqualError(t, err, fmt.Sprintf("read tcp %s->%s: i/o timeout", conn.LocalAddr(), conn.RemoteAddr()))
	}()

	// Test dial auth success.
	func() {
		conn, err := Dial(server.Addr(), MockPassword)
		if assert.NoError(t, err) {
			// Close connection if established.
			assert.NoError(t, conn.Close())
		}
	}()
}

func TestConn_Execute(t *testing.T) {
	server, err := NewMockServer()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		assert.NoError(t, server.Close())
		close(server.errors)
		for err := range server.errors {
			assert.NoError(t, err)
		}
	}()

	// Test incorrect command.
	func() {
		conn, err := Dial(server.Addr(), MockPassword)
		if !assert.NoError(t, err) {
			return
		}
		defer assert.NoError(t, conn.Close())

		result, err := conn.Execute("")
		assert.Equal(t, err, ErrCommandEmpty)
		assert.Equal(t, 0, len(result))

		result, err = conn.Execute(string(make([]byte, 1001)))
		assert.Equal(t, err, ErrCommandTooLong)
		assert.Equal(t, 0, len(result))
	}()

	// Test use of closed network connection.
	func() {
		conn, err := Dial(server.Addr(), MockPassword)
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, conn.Close())

		result, err := conn.Execute(MockCommandHelp)
		assert.EqualError(t, err, fmt.Sprintf("write tcp %s->%s: use of closed network connection", conn.LocalAddr(), conn.RemoteAddr()))
		assert.Equal(t, 0, len(result))
	}()

	// Test read deadline.
	func() {
		conn, err := Dial(server.Addr(), MockPassword, SetDeadline(1*time.Second))
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		result, err := conn.Execute("deadline")
		assert.EqualError(t, err, fmt.Sprintf("read tcp %s->%s: i/o timeout", conn.LocalAddr(), conn.RemoteAddr()))
		assert.Equal(t, 0, len(result))
	}()

	// Test success command help.
	func() {
		conn, err := Dial(server.Addr(), MockPassword)
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		result, err := conn.Execute(MockCommandHelp)
		assert.NoError(t, err)
		assert.Equal(t, len([]byte(MockCommandHelpResponse)), len(result))
		assert.Equal(t, MockCommandHelpResponse, result)
	}()

	// Test Rust.
	func() {
		conn, err := Dial(server.Addr(), MockPassword, SetDeadline(1*time.Second))
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		result, err := conn.Execute("rust")
		assert.NoError(t, err)
		assert.Equal(t, len([]byte("rust")), len(result))
		assert.Equal(t, "rust", result)
	}()
}
