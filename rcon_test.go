package rcon_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorcon/rcon"
	"github.com/gorcon/rcon/rcontest"
	"github.com/stretchr/testify/assert"
)

func authHandler(c *rcontest.Context) {
	switch c.Request().Body() {
	case c.Server().Settings.Password:
		rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, "").WriteTo(c.Conn())
		rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, c.Request().ID, "").WriteTo(c.Conn())
	case "invalid packet type":
		rcon.NewPacket(42, c.Request().ID, "").WriteTo(c.Conn())
	case "another":
		rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, 42, "").WriteTo(c.Conn())
	default:
		rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, -1, string([]byte{0x00})).WriteTo(c.Conn())
	}
}

func commandHandler(c *rcontest.Context) {
	writeWithInvalidPadding := func(conn io.Writer, packet *rcon.Packet) {
		buffer := bytes.NewBuffer(make([]byte, 0, packet.Size+4))

		binary.Write(buffer, binary.LittleEndian, packet.Size)
		binary.Write(buffer, binary.LittleEndian, packet.ID)
		binary.Write(buffer, binary.LittleEndian, packet.Type)

		// Write command body, null terminated ASCII string and an empty ASCIIZ string.
		// Second padding byte is incorrect.
		buffer.Write(append([]byte(packet.Body()), 0x00, 0x01))

		buffer.WriteTo(conn)
	}

	switch c.Request().Body() {
	case "help":
		responseBody := "lorem ipsum dolor sit amet"
		rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, responseBody).WriteTo(c.Conn())
	case "rust":
		// Write specific Rust package.
		rcon.NewPacket(4, c.Request().ID, "").WriteTo(c.Conn())

		rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, -1, c.Request().Body()).WriteTo(c.Conn())
	case "padding":
		writeWithInvalidPadding(c.Conn(), rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, ""))
	case "another":
		rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, 42, "").WriteTo(c.Conn())
	default:
		rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, "unknown command").WriteTo(c.Conn())
	}
}

func TestDial(t *testing.T) {
	server := rcontest.NewServer(rcontest.SetSettings(rcontest.Settings{Password: "password"}))
	defer server.Close()

	t.Run("connection refused", func(t *testing.T) {
		wantErrContains := "connect: connection refused"

		_, err := rcon.Dial("127.0.0.2:12345", "password")
		if err == nil || !strings.Contains(err.Error(), wantErrContains) {
			t.Errorf("got err %q, want to contain %q", err, wantErrContains)
		}
	})

	t.Run("connection timeout", func(t *testing.T) {
		server := rcontest.NewServer(rcontest.SetSettings(rcontest.Settings{Password: "password", AuthResponseDelay: 6 * time.Second}))
		defer server.Close()

		wantErrContains := "i/o timeout"

		_, err := rcon.Dial(server.Addr(), "", rcon.SetDialTimeout(5*time.Second))
		if err == nil || !strings.Contains(err.Error(), wantErrContains) {
			t.Errorf("got err %q, want to contain %q", err, wantErrContains)
		}
	})

	t.Run("authentication failed", func(t *testing.T) {
		_, err := rcon.Dial(server.Addr(), "wrong")
		if !errors.Is(err, rcon.ErrAuthFailed) {
			t.Errorf("got err %q, want %q", err, rcon.ErrAuthFailed)
		}
	})

	t.Run("invalid packet type", func(t *testing.T) {
		server := rcontest.NewServer(
			rcontest.SetSettings(rcontest.Settings{Password: "password"}),
			rcontest.SetAuthHandler(authHandler),
		)
		defer server.Close()

		_, err := rcon.Dial(server.Addr(), "invalid packet type")
		if !errors.Is(err, rcon.ErrInvalidAuthResponse) {
			t.Errorf("got err %q, want %q", err, rcon.ErrInvalidAuthResponse)
		}
	})

	t.Run("invalid response id", func(t *testing.T) {
		server := rcontest.NewServer(
			rcontest.SetSettings(rcontest.Settings{Password: "password"}),
			rcontest.SetAuthHandler(authHandler),
		)
		defer server.Close()

		_, err := rcon.Dial(server.Addr(), "another")
		if !errors.Is(err, rcon.ErrInvalidPacketID) {
			t.Errorf("got err %q, want %q", err, rcon.ErrInvalidPacketID)
		}
	})

	t.Run("auth success", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if err != nil {
			t.Errorf("got err %q, want %v", err, nil)
			return
		}

		conn.Close()
	})
}

func TestConn_Execute(t *testing.T) {
	server := rcontest.NewUnstartedServer()
	server.Settings.Password = "password"
	server.SetCommandHandler(commandHandler)
	server.Start()
	defer server.Close()

	t.Run("incorrect command", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if !assert.NoError(t, err) {
			return
		}
		defer assert.NoError(t, conn.Close())

		result, err := conn.Execute("")
		assert.Equal(t, err, rcon.ErrCommandEmpty)
		assert.Equal(t, 0, len(result))

		result, err = conn.Execute(string(make([]byte, 1001)))
		assert.Equal(t, err, rcon.ErrCommandTooLong)
		assert.Equal(t, 0, len(result))
	})

	t.Run("closed network connection 1", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password", rcon.SetDeadline(0))
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, conn.Close())

		result, err := conn.Execute("help")
		assert.EqualError(t, err, fmt.Sprintf("write tcp %s->%s: use of closed network connection", conn.LocalAddr(), conn.RemoteAddr()))
		assert.Equal(t, 0, len(result))
	})

	t.Run("closed network connection 2", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, conn.Close())

		result, err := conn.Execute("help")
		assert.EqualError(t, err, fmt.Sprintf("set tcp %s: use of closed network connection", conn.LocalAddr()))
		assert.Equal(t, 0, len(result))
	})

	t.Run("read deadline", func(t *testing.T) {
		server := rcontest.NewServer(rcontest.SetSettings(rcontest.Settings{Password: "password", CommandResponseDelay: 2 * time.Second}))
		defer server.Close()

		conn, err := rcon.Dial(server.Addr(), "password", rcon.SetDeadline(1*time.Second))
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

	t.Run("invalid padding", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		_, err = conn.Execute("padding")
		assert.Equal(t, rcon.ErrInvalidPacketPadding, err)
	})

	t.Run("invalid response id", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		_, err = conn.Execute("another")
		assert.Equal(t, rcon.ErrInvalidPacketID, err)
	})

	t.Run("success help command", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		result, err := conn.Execute("help")
		assert.NoError(t, err)

		assert.Equal(t, "lorem ipsum dolor sit amet", result)
	})

	t.Run("rust workaround", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password", rcon.SetDeadline(1*time.Second))
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		result, err := conn.Execute("rust")
		assert.NoError(t, err)

		assert.Equal(t, "rust", result)
	})

	if run := getVar("TEST_PZ_SERVER", "false"); run == "true" {
		addr := getVar("TEST_PZ_SERVER_ADDR", "127.0.0.1:16260")
		password := getVar("TEST_PZ_SERVER_PASSWORD", "docker")

		t.Run("pz server", func(t *testing.T) {
			needle := func() string {
				n := `List of server commands :
* addalltowhitelist : Add all the current users connected with a password in the whitelist, so their account is protected.
* additem : Add an item to a player, if no username is given the item will be added to you, count is optional, use /additem \"username\" \"module.item\" count, ex : /additem \"rj\" \"Base.Axe\" count
* adduser : Use this command to add a new user in a whitelisted server, use : /adduser \"username\" \"pwd\"
* addusertowhitelist : Add the user connected with a password in the whitelist, so his account is protected, use : /addusertowhitelist \"username\"
* addvehicle : Spawn a new vehicle, use: /addvehicle \"script\" \"user or x,y,z\", ex /addvehicle \"Base.VanAmbulance\" \"rj\"
* addxp : Add experience points to a player, use : /addxp \"playername\" perkname=xp, ex /addxp \"rj\" Woodwork=2
* alarm : Sound a building alarm at the admin's position.  Must be in a room.
* banid : Ban a SteamID, use : /banid SteamID
* banuser : Ban a user, add a -ip to also ban his ip, add a -r \"reason\" to specify a reason for the ban, use : /banuser \"username\" -ip -r \"reason\", ex /banuser \"rj\" -ip -r \"spawn kill\"
* changeoption : Use this to change a server option, use : /changeoption optionName \"newValue\"
* chopper : Start the choppers (do noise on a random player)
* createhorde : Use this to spawn a horde near a player, use : /createhorde count \"username\", ex /createhorde 150 \"rj\", username is optional except from the server console.
* godmod : Set a player invincible, if no username set it make you invincible, if no value it toggle it, use : /godmode \"username\" -value, ex /godmode \"rj\" -true (could be -false)
* gunshot : Start a gunshot (do noise on a random player)
* help : Help
* invisible : Set a player invisible zombie will ignore him, if no username set it make you invisible, if no value it toggle it, use : /invisible \"username\" -value, ex /invisible \"rj\" -true (could be -false)
* kickuser : Kick a user, add a -r \"reason\" to specify a reason for the kick, use : /kickuser \"username\" -r \"reason\"
* noclip : A player with noclip won't collide on anything, if no value it toggle it, use : /noclip \"username\" -value, ex /noclip \"rj\" -true (could be -false)
* players : List the players connected
* quit : Quit the server (but save it before)
* releasesafehouse : Release a safehouse you are the owner of, use : /releasesafehouse
* reloadlua : Reload a Lua script, use : /reloadlua \"filename\"
* reloadoptions : Reload the options on the server (ServerOptions.ini) and send them to the clients
* removeuserfromwhitelist : Remove the user from the whitelist, use: /removeuserfromwhitelist \"username\"
* save : Save the current world
* sendpulse : Toggle sending server performance info to this client, use : /sendpulse
* servermsg : Use this to broadcast a message to all connected players, use : /servermsg my message !
* setaccesslevel : Use it to set new access level to a player, acces level: admin, moderator, overseer, gm, observer. use : /setaccesslevel \"username\" \"accesslevel\", ex: /setaccesslevel \"rj\" \"moderator\"
* showoptions : Show the list of current Server options with their values.
* startrain : Start rain on the server
* stoprain : Stop rain on the server
* teleport : Teleport to a player, once teleported, wait 2 seconds to show map, use : /teleport \"playername\" or /teleport \"player1\" \"player2\", ex /teleport \"rj\" or /teleport \"rj\" \"toUser\"
* teleportto : Teleport to coordinates, use: /teleportto x,y,z, ex /teleportto 100098,189980,0
* unbanid : Unban a SteamID, use : /unbanid SteamID
* unbanuser : Unban a player, use : /unbanuser \"username\"
* voiceban : Block voice from user \"username\", use : /voiceban \"username\" -value, ex /voiceban \"rj\" -true (could be -false)`

				n = strings.Replace(n, "List of server commands :", "List of server commands : ", -1)

				return n
			}()

			conn, err := rcon.Dial(addr, password)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				assert.NoError(t, conn.Close())
			}()

			result, err := conn.Execute("help")
			assert.NoError(t, err)

			assert.Equal(t, needle, result)
		})
	}

	if run := getVar("TEST_RUST_SERVER", "false"); run == "true" {
		addr := getVar("TEST_RUST_SERVER_ADDR", "127.0.0.1:28016")
		password := getVar("TEST_RUST_SERVER_PASSWORD", "docker")

		t.Run("rust server", func(t *testing.T) {
			conn, err := rcon.Dial(addr, password)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				assert.NoError(t, conn.Close())
			}()

			result, err := conn.Execute("status")
			assert.NoError(t, err)
			assert.NotEmpty(t, result)

			fmt.Println(result)
		})
	}
}

// getVar returns environment variable or default value.
func getVar(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
