package rcon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteConsole_Execute(t *testing.T) {
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

	conn, err := Dial(server.Addr(), MockPassword)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		assert.NoError(t, conn.Close())
	}()

	result, err := conn.Execute(MockCommandHelp)
	assert.NoError(t, err)
	assert.Equal(t, len([]byte(MockCommandHelpResponse)), len(result))
	assert.Equal(t, MockCommandHelpResponse, result)
}
