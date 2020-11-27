package rcon

import (
	"bytes"
	"testing"
)

func TestNewPacket(t *testing.T) {
	body := []byte("testdata")
	packet := NewPacket(SERVERDATA_RESPONSE_VALUE, 42, string(body))

	if packet.Body() != string(body) {
		t.Errorf("%q, want %q", packet.Body(), body)
	}

	want := int32(len([]byte(body))) + PacketHeaderSize + PacketPaddingSize
	if packet.Size != want {
		t.Errorf("got %d, want %d", packet.Size, want)
	}
}

func TestPacket_WriteTo(t *testing.T) {
	t.Run("check bytes written", func(t *testing.T) {
		body := []byte("testdata")
		packet := NewPacket(SERVERDATA_RESPONSE_VALUE, 42, string(body))

		var buffer bytes.Buffer
		n, err := packet.WriteTo(&buffer)
		if err != nil {
			t.Fatal(err)
		}

		wantN := packet.Size + 4
		if n != int64(wantN) {
			t.Errorf("got %d, want %d", n, int64(wantN))
		}
	})
}
