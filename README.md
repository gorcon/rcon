# rcon
Source RCON Protocol implementation in Go.

## Protocol Specifications

RCON Protocol is described in the [valve documentation](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol)

## Supported Games

* [Project Zomboid](https://store.steampowered.com/app/108600) 
* [Conan Exiles](https://store.steampowered.com/app/440900)
* [Minecraft](https://www.minecraft.net)

Open pull request if you have successfully used a package with another game with rcon support and add it to the list.

## Install

```text
go get github.com/gorcon/rcon
```

Or use dependency manager such as dep or vgo.

## Usage

```go
package main

import (
	"log"
	"fmt"

	"github.com/gorcon/rcon"
)

func main() {
	conn, err := rcon.Dial("127.0.01:16260", "password")
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

Go 1.11 or higher

## Contribute

Contributions are more than welcome! 

If you think that you have found a bug, create an issue and publish the minimum amount of code triggering the bug so 
it can be reproduced.

If you want to fix the bug then you can create a pull request. If possible, write a test that will cover this bug.

## License

MIT License, see [LICENSE](LICENSE)
