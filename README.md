# Rcon
[![GitHub Build](https://github.com/gorcon/rcon/workflows/build/badge.svg)](https://github.com/gorcon/rcon/actions)
[![Go Coverage](https://github.com/gorcon/rcon/wiki/coverage.svg)](https://raw.githack.com/wiki/gorcon/rcon/coverage.html)
[![Go Report Card](https://goreportcard.com/badge/github.com/gorcon/rcon)](https://goreportcard.com/report/github.com/gorcon/rcon)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/gorcon/rcon)

Source RCON Protocol implementation in Go.

## Protocol Specifications
RCON Protocol described in the [valve documentation](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol).

## Supported Games
* [Project Zomboid](https://store.steampowered.com/app/108600) 
* [Conan Exiles](https://store.steampowered.com/app/440900)
* [Rust](https://store.steampowered.com/app/252490) (add +rcon.web 0 to the args when starting the server)
* [ARK: Survival Evolved](https://store.steampowered.com/app/346110)
* [Counter-Strike: Global Offensive](https://store.steampowered.com/app/730)
* [Minecraft](https://www.minecraft.net)
* [Palworld](https://store.steampowered.com/app/1623730/Palworld/)
* [Factorio](https://www.factorio.com/) (start the server with `--rcon-bind`/`--rcon-port` and `--rcon-password` args)

Open pull request if you have successfully used a package with another game with rcon support and add it to the list.

## Install
```text
go get github.com/gorcon/rcon
```

See [Changelog](CHANGELOG.md) for release details.

## Usage
```go
package main

import (
	"fmt"
	"log"

	"github.com/gorcon/rcon"
)

func main() {
	conn, err := rcon.Dial("127.0.0.1:16260", "password")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	response, err := conn.Execute("help")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(response)	
}
```

### With an existing net.Conn
If you wish to initialize a RCON connection with an already initialized net.Conn, you can use the `Open` function:
```go
package main

import (
	"fmt"
	"log"
	"net"

	"github.com/gorcon/rcon"
)

func main() {
	netConn, err := net.Dial("tcp", "127.0.0.1:16260")
	if err != nil {
		// Failed to open TCP connection to the server.
		log.Fatalf("expected nil got error: %s", err)
	}
	defer netConn.Close()
	
	conn, err := rcon.Open(netConn, "password")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	response, err := conn.Execute("help")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(response)	
}
```

## Requirements
Go 1.15 or higher

## Contribute
Contributions are more than welcome! 

If you think that you have found a bug, create an issue and publish the minimum amount of code triggering the bug, so 
it can be reproduced.

If you want to fix the bug then you can create a pull request. If possible, write a test that will cover this bug.

## License
MIT License, see [LICENSE](LICENSE)
