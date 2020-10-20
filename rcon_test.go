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

	t.Run("connection refused", func(t *testing.T) {
		conn, err := Dial("127.0.0.2:12345", MockPassword)
		if !assert.Error(t, err) {
			// Close connection if established.
			assert.NoError(t, conn.Close())
		}
		assert.EqualError(t, err, "dial tcp 127.0.0.2:12345: connect: connection refused")
	})

	t.Run("connection timeout", func(t *testing.T) {
		conn, err := Dial(server.Addr(), "timeout", SetDialTimeout(5*time.Second))
		if !assert.Error(t, err) {
			assert.NoError(t, conn.Close())
		}
		assert.EqualError(t, err, fmt.Sprintf("read tcp %s->%s: i/o timeout", conn.LocalAddr(), conn.RemoteAddr()))
	})

	t.Run("authentication failed", func(t *testing.T) {
		conn, err := Dial(server.Addr(), "wrong")
		if !assert.Error(t, err) {
			assert.NoError(t, conn.Close())
		}
		assert.EqualError(t, err, "authentication failed")
	})

	t.Run("auth success", func(t *testing.T) {
		conn, err := Dial(server.Addr(), MockPassword)
		if assert.NoError(t, err) {
			assert.NoError(t, conn.Close())
		}
	})
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

	t.Run("incorrect command", func(t *testing.T) {
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
	})

	t.Run("closed network connection 1", func(t *testing.T) {
		conn, err := Dial(server.Addr(), MockPassword, SetDeadline(0))
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, conn.Close())

		result, err := conn.Execute(MockCommandHelp)
		assert.EqualError(t, err, fmt.Sprintf("write tcp %s->%s: use of closed network connection", conn.LocalAddr(), conn.RemoteAddr()))
		assert.Equal(t, 0, len(result))
	})

	t.Run("closed network connection 2", func(t *testing.T) {
		conn, err := Dial(server.Addr(), MockPassword)
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, conn.Close())

		result, err := conn.Execute(MockCommandHelp)
		assert.EqualError(t, err, fmt.Sprintf("set tcp %s: use of closed network connection", conn.LocalAddr()))
		assert.Equal(t, 0, len(result))
	})

	t.Run("read deadline", func(t *testing.T) {
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
	})

	t.Run("success help command", func(t *testing.T) {
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
	})

	t.Run("rust workaround", func(t *testing.T) {
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
	})
}
